package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claudeception",
	Short: "Claudeception agent for managing CLAUDE.md configurations",
	Long: `Claudeception is a local agent that syncs CLAUDE.md configurations
from a central server to your development environment.

It watches for project changes, detects context, and applies
the appropriate rules to maintain consistent Claude behavior.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("claudeception agent v0.1.0")
	},
}
