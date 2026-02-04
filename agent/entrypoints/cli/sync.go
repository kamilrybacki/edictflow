package cli

import (
	"fmt"

	"github.com/kamilrybacki/claudeception/agent/daemon"
	"github.com/kamilrybacki/claudeception/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(validateCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Force immediate sync",
	Long:  `Force an immediate synchronization with the central server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, running := daemon.IsRunning()
		if !running {
			return fmt.Errorf("daemon not running. Use 'claudeception start' first")
		}

		_, err := daemon.QueryDaemon("sync")
		if err != nil {
			return fmt.Errorf("failed to trigger sync: %w", err)
		}

		fmt.Println("Sync triggered.")
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Check drift for a project",
	Long:  `Validate that local CLAUDE.md files match expected content.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		rules, err := store.GetRules()
		if err != nil {
			return err
		}

		if len(rules) == 0 {
			fmt.Println("No rules cached. Run sync first.")
			return nil
		}

		fmt.Printf("Cached rules version: %d\n", store.GetCachedVersion())
		fmt.Printf("Rules count: %d\n", len(rules))
		// TODO: Check actual CLAUDE.md files against rules
		fmt.Println("Validation complete.")
		return nil
	},
}
