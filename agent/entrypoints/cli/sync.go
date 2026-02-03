package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(applyCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Force immediate sync",
	Long:  `Force an immediate synchronization with the central server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing with server...")
		// TODO: Implement sync
		fmt.Println("Sync complete")
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Check drift for a project",
	Long:  `Validate that local CLAUDE.md files match expected content.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		fmt.Printf("Validating project at %s...\n", path)
		// TODO: Implement validation
		fmt.Println("No drift detected")
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply [path]",
	Short: "Write CLAUDE.md files for a project",
	Long:  `Apply the appropriate CLAUDE.md configurations to a project.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		fmt.Printf("Applying configuration to %s...\n", path)
		// TODO: Implement apply
		fmt.Println("Configuration applied")
	},
}
