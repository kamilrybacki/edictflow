package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status",
	Long:  `Show the current connection status, cached config age, and active projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Status: Disconnected")
		fmt.Println("Cached config: None")
		fmt.Println("Active projects: 0")
		// TODO: Implement actual status check
	},
}
