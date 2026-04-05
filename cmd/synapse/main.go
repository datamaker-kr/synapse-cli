package main

import (
	"os"

	"github.com/datamaker-kr/synapse-cli/internal/cmd"
	mcpserver "github.com/datamaker-kr/synapse-cli/internal/mcp"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	mcpserver.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
