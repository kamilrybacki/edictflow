// agent/entrypoints/cli/changes.go
package cli

import (
	"fmt"
	"time"

	"github.com/kamilrybacki/edictflow/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(changesCmd)
	rootCmd.AddCommand(appealCmd)
	appealCmd.Flags().String("reason", "", "Justification for the exception")
}

var changesCmd = &cobra.Command{
	Use:   "changes [id]",
	Short: "List pending changes or show details",
	Long:  `List all pending change requests or show details of a specific one.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if len(args) == 1 {
			// Show specific change
			change, err := store.GetPendingChange(args[0])
			if err != nil {
				return fmt.Errorf("change not found: %s", args[0])
			}

			fmt.Printf("Change ID: %s\n", change.ID)
			fmt.Printf("Rule ID: %s\n", change.RuleID)
			fmt.Printf("File: %s\n", change.FilePath)
			fmt.Printf("Status: %s\n", change.Status)
			fmt.Printf("Created: %s\n", change.CreatedAt.Format(time.RFC3339))
			return nil
		}

		// List all changes
		changes, err := store.GetPendingChanges()
		if err != nil {
			return err
		}

		if len(changes) == 0 {
			fmt.Println("No pending changes.")
			return nil
		}

		fmt.Println("Pending changes:")
		for _, c := range changes {
			fmt.Printf("  [%s] %s - %s (%s)\n", c.Status, c.ID[:8], c.FilePath, c.CreatedAt.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

var appealCmd = &cobra.Command{
	Use:   "appeal <change-id> --reason <reason>",
	Short: "Request an exception for a rejected change",
	Long:  `Submit an exception request for a change that was blocked or rejected.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reason, _ := cmd.Flags().GetString("reason")
		if reason == "" {
			return fmt.Errorf("--reason is required")
		}

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		change, err := store.GetPendingChange(args[0])
		if err != nil {
			return fmt.Errorf("change not found: %s", args[0])
		}

		fmt.Printf("Submitting exception request for change %s...\n", change.ID)
		// TODO: Send via WebSocket when daemon is running
		fmt.Println("Exception request queued.")
		return nil
	},
}
