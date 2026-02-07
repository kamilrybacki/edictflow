// agent/entrypoints/cli/login.go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/kamilrybacki/edictflow/agent/auth"
	"github.com/kamilrybacki/edictflow/agent/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("server", "s", "", "API Server URL (default: $EDICTFLOW_API_SERVER or http://localhost:8080)")
	loginCmd.Flags().StringP("ws-server", "w", "", "WebSocket Server URL for daemon (default: $EDICTFLOW_SERVER or same as --server)")
	loginCmd.Flags().StringP("email", "e", "", "Email for credentials login")
	loginCmd.Flags().StringP("password", "p", "", "Password for credentials login (insecure, prefer interactive prompt)")
	loginCmd.Flags().Bool("device", false, "Use device code flow instead of credentials")
}

var loginCmd = &cobra.Command{
	Use:   "login [server-url]",
	Short: "Authenticate with central server",
	Long: `Authenticate with the Edictflow server.

By default, uses email/password credentials (interactive prompt).
Use --device flag for device code flow (opens browser).

Examples:
  edictflow-agent login http://server:8080
  edictflow-agent login -s http://server:8080 -e user@example.com
  edictflow-agent login --device http://server:8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		wsServerURL, _ := cmd.Flags().GetString("ws-server")

		// Allow positional argument to override flag
		if len(args) > 0 {
			serverURL = args[0]
		}

		// Use environment variables as defaults
		if serverURL == "" {
			serverURL = os.Getenv("EDICTFLOW_API_SERVER")
		}
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		if wsServerURL == "" {
			wsServerURL = os.Getenv("EDICTFLOW_SERVER")
		}
		if wsServerURL == "" {
			wsServerURL = serverURL // Default to same as API server
		}

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if store.IsLoggedIn() {
			authInfo, _ := store.GetAuth()
			fmt.Printf("Already logged in as %s (%s)\n", authInfo.UserName, authInfo.UserEmail)
			fmt.Println("Use 'edictflow-agent logout' to sign out first.")
			return nil
		}

		// Save server URLs for future use
		if err := store.SaveServerURL(serverURL); err != nil {
			return fmt.Errorf("failed to save server URL: %w", err)
		}
		if err := store.SaveWSServerURL(wsServerURL); err != nil {
			return fmt.Errorf("failed to save WebSocket server URL: %w", err)
		}

		useDevice, _ := cmd.Flags().GetBool("device")
		if useDevice {
			return loginWithDeviceFlow(serverURL, store)
		}

		return loginWithCredentials(cmd, serverURL, store)
	},
}

func loginWithCredentials(cmd *cobra.Command, serverURL string, store *storage.Storage) error {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	// Prompt for email if not provided
	if email == "" {
		fmt.Print("Email: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(input)
	}

	// Prompt for password if not provided
	if password == "" {
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // newline after password
		password = string(bytePassword)
	}

	if email == "" || password == "" {
		return fmt.Errorf("email and password are required")
	}

	fmt.Println("Authenticating...")

	client := auth.NewCredentialsClient(serverURL)
	resp, err := client.Login(email, password)
	if err != nil {
		return err
	}

	// Calculate expiry
	expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	if resp.ExpiresIn == 0 {
		// Default to 24 hours if not specified
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	teamID := ""
	if resp.User.TeamID != nil {
		teamID = *resp.User.TeamID
	}

	authInfo := storage.AuthInfo{
		AccessToken: resp.Token,
		ExpiresAt:   expiresAt,
		UserID:      resp.User.ID,
		UserEmail:   resp.User.Email,
		UserName:    resp.User.Name,
		TeamID:      teamID,
	}

	if err := store.SaveAuth(authInfo); err != nil {
		return fmt.Errorf("failed to save auth: %w", err)
	}

	fmt.Printf("\nLogin successful! Welcome, %s\n", resp.User.Name)
	return nil
}

func loginWithDeviceFlow(serverURL string, store *storage.Storage) error {
	client := auth.NewDeviceFlowClient(serverURL)

	fmt.Println("Initiating device authorization...")
	deviceAuth, err := client.InitiateDeviceAuth()
	if err != nil {
		return fmt.Errorf("failed to initiate auth: %w", err)
	}

	fmt.Println()
	fmt.Println("To authorize this device:")
	fmt.Printf("  1. Open: %s\n", deviceAuth.VerificationURI)
	fmt.Printf("  2. Enter code: %s\n", deviceAuth.UserCode)
	fmt.Println()

	// Try to open browser
	openBrowser(deviceAuth.VerificationURI + "?code=" + deviceAuth.UserCode)

	fmt.Println("Waiting for authorization...")
	token, err := client.PollForToken(deviceAuth.DeviceCode, deviceAuth.Interval, deviceAuth.ExpiresIn)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	authInfo := storage.AuthInfo{
		AccessToken: token.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		UserID:      "pending",
		UserEmail:   "pending",
		UserName:    "User",
	}

	if err := store.SaveAuth(authInfo); err != nil {
		return fmt.Errorf("failed to save auth: %w", err)
	}

	fmt.Println()
	fmt.Println("Login successful!")
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
