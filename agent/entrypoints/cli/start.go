package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the Claudeception daemon to sync configurations in the background.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Claudeception daemon...")
		// TODO: Implement daemon start
		fmt.Println("Daemon started successfully")
	},
}
