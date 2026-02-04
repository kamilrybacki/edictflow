// agent/entrypoints/cli/stop.go
package cli

import (
	"github.com/kamilrybacki/claudeception/agent/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  `Stop the running Claudeception daemon.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Stop()
	},
}
