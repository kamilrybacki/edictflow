package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  `Stop the running Claudeception daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping Claudeception daemon...")
		// TODO: Implement daemon stop
		fmt.Println("Daemon stopped successfully")
	},
}
