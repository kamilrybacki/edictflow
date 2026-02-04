// agent/entrypoints/cli/start.go
package cli

import (
	"fmt"

	"github.com/kamilrybacki/claudeception/agent/daemon"
	"github.com/kamilrybacki/claudeception/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Bool("foreground", false, "Run in foreground (don't fork)")
	startCmd.Flags().StringP("server", "s", "http://localhost:8080", "Server URL")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the Claudeception daemon to sync configurations in the background.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		foreground, _ := cmd.Flags().GetBool("foreground")
		serverURL, _ := cmd.Flags().GetString("server")

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if !store.IsLoggedIn() {
			return fmt.Errorf("not logged in. Run 'claudeception login' first")
		}

		return daemon.Start(serverURL, foreground)
	},
}
