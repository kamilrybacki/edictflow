package main

import (
	"os"

	"github.com/kamilrybacki/claudeception/agent/entrypoints/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
