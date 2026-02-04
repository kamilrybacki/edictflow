// agent/entrypoints/cli/logout.go
package cli

import (
	"fmt"

	"github.com/kamilrybacki/claudeception/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out and clear credentials",
	Long:  `Clear saved authentication credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if !store.IsLoggedIn() {
			fmt.Println("Not logged in.")
			return nil
		}

		if err := store.ClearAuth(); err != nil {
			return fmt.Errorf("failed to clear auth: %w", err)
		}

		fmt.Println("Logged out successfully.")
		return nil
	},
}
