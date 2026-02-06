package main

import (
	"os"

	"github.com/kamilrybacki/edictflow/agent/entrypoints/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
