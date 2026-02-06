// agent/entrypoints/cli/watch.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kamilrybacki/edictflow/agent/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(unwatchCmd)
	rootCmd.AddCommand(listCmd)
}

var watchCmd = &cobra.Command{
	Use:   "watch <path>",
	Short: "Add a project to watch list",
	Long:  `Start monitoring a project for CLAUDE.md changes.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("path not found: %s", path)
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", path)
		}

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if err := store.AddProject(path); err != nil {
			return fmt.Errorf("failed to add project: %w", err)
		}

		fmt.Printf("Now watching: %s\n", path)
		fmt.Println("Restart daemon to pick up new projects.")
		return nil
	},
}

var unwatchCmd = &cobra.Command{
	Use:   "unwatch <path>",
	Short: "Remove a project from watch list",
	Long:  `Stop monitoring a project.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		if err := store.RemoveProject(path); err != nil {
			return fmt.Errorf("failed to remove project: %w", err)
		}

		fmt.Printf("Stopped watching: %s\n", path)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List watched projects",
	Long:  `Show all projects being monitored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		projects, err := store.GetProjects()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Println("No projects being watched.")
			fmt.Println("Use 'edictflow watch <path>' to add one.")
			return nil
		}

		fmt.Println("Watched projects:")
		for _, p := range projects {
			fmt.Printf("  %s\n", p.Path)
			if len(p.DetectedContext) > 0 {
				fmt.Printf("    Context: %v\n", p.DetectedContext)
			}
			if len(p.DetectedTags) > 0 {
				fmt.Printf("    Tags: %v\n", p.DetectedTags)
			}
		}
		return nil
	},
}
