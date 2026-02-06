// agent/entrypoints/cli/status.go
package cli

import (
	"encoding/json"
	"fmt"

	"github.com/kamilrybacki/edictflow/agent/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status",
	Long:  `Show the current connection status, cached config age, and active projects.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, running := daemon.IsRunning()
		if !running {
			fmt.Println("Status: Daemon not running")
			fmt.Println("Run 'edictflow start' to start the daemon")
			return nil
		}

		fmt.Printf("Status: Daemon running (PID %d)\n", pid)

		data, err := daemon.QueryDaemon("status")
		if err != nil {
			fmt.Println("Could not query daemon")
			return nil
		}

		var status daemon.StatusResponse
		if err := json.Unmarshal(data, &status); err != nil {
			return nil
		}

		if status.Connected {
			fmt.Println("Server: Connected")
		} else {
			fmt.Println("Server: Disconnected")
		}

		fmt.Printf("Cached config version: %d\n", status.CachedVersion)
		fmt.Printf("Watched projects: %d\n", len(status.Projects))
		for _, p := range status.Projects {
			fmt.Printf("  - %s\n", p)
		}

		if status.PendingMsgs > 0 {
			fmt.Printf("Pending messages: %d\n", status.PendingMsgs)
		}

		return nil
	},
}
