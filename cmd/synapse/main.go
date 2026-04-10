package main

import (
	"os"
	"runtime/debug"

	"github.com/datamaker-kr/synapse-cli/internal/cmd"
	mcpserver "github.com/datamaker-kr/synapse-cli/internal/mcp"
)

// version is set via ldflags at build time.
// Fallback chain: ldflags → runtime/debug module version → "dev"
var version = "dev"

func resolveVersion() string {
	// 1. ldflags로 주입된 값 (make build, goreleaser)
	if version != "dev" {
		return version
	}
	// 2. go install ...@vX.Y.Z 시 모듈 버전 자동 주입
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	// 3. 개발 중 fallback
	return "dev"
}

func main() {
	v := resolveVersion()
	cmd.SetVersion(v)
	mcpserver.SetVersion(v)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
