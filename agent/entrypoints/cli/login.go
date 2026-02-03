package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with central server",
	Long:  `Open browser to authenticate with the central server via OAuth.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Opening browser for authentication...")
		// TODO: Implement OAuth flow
		fmt.Println("Login successful")
	},
}
