// agent/entrypoints/cli/login.go
package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/kamilrybacki/claudeception/agent/auth"
	"github.com/kamilrybacki/claudeception/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("server", "s", "http://localhost:8080", "Server URL")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with central server",
	Long:  `Authenticate using device code flow. Opens browser for authorization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if store.IsLoggedIn() {
			authInfo, _ := store.GetAuth()
			fmt.Printf("Already logged in as %s (%s)\n", authInfo.UserName, authInfo.UserEmail)
			fmt.Println("Use 'claudeception logout' to sign out first.")
			return nil
		}

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

		// Save token (we don't have user info yet, will fetch on next command)
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
	},
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
