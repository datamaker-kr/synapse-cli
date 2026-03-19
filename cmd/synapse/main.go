package main

import (
	"os"

	"github.com/datamaker-kr/synapse-cli/internal/cmd"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
