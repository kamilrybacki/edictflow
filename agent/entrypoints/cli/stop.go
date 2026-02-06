// agent/entrypoints/cli/stop.go
package cli

import (
	"github.com/kamilrybacki/edictflow/agent/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  `Stop the running Edictflow daemon.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Stop()
	},
}
