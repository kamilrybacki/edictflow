// agent/entrypoints/cli/start.go
package cli

import (
	"fmt"
	"os"

	"github.com/kamilrybacki/edictflow/agent/daemon"
	"github.com/kamilrybacki/edictflow/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Bool("foreground", false, "Run in foreground (don't fork)")
	startCmd.Flags().StringP("server", "s", "", "WebSocket Server URL (default: saved from login or $EDICTFLOW_SERVER)")
	startCmd.Flags().Duration("poll-interval", 0, "File watcher poll interval (e.g., 500ms). Use for container environments where fsnotify is unreliable.")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the Edictflow daemon to sync configurations in the background.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		foreground, _ := cmd.Flags().GetBool("foreground")
		serverURL, _ := cmd.Flags().GetString("server")
		pollInterval, _ := cmd.Flags().GetDuration("poll-interval")

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if !store.IsLoggedIn() {
			return fmt.Errorf("not logged in. Run 'edictflow login' first")
		}

		// Determine WebSocket server URL
		if !cmd.Flags().Changed("server") {
			// First try saved WebSocket URL from login
			savedURL, err := store.GetWSServerURL()
			if err == nil && savedURL != "" {
				serverURL = savedURL
			} else {
				// Fall back to environment variable
				serverURL = os.Getenv("EDICTFLOW_SERVER")
			}
			// Finally fall back to saved API server URL
			if serverURL == "" {
				savedURL, err := store.GetServerURL()
				if err == nil && savedURL != "" {
					serverURL = savedURL
				}
			}
		}

		if serverURL == "" {
			return fmt.Errorf("no server URL configured. Use --server flag or set EDICTFLOW_SERVER")
		}

		return daemon.Start(serverURL, foreground, pollInterval)
	},
}
