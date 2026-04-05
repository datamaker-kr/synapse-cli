package cmd

import (
	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/config"
	mcpserver "github.com/datamaker-kr/synapse-cli/internal/mcp"
)

func newMcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server (stdio) for AI assistant integration",
		Long: `MCP(Model Context Protocol) 서버를 stdio 모드로 실행한다.
Claude Code, Claude Desktop 등 MCP 클라이언트와 연동하여
Synapse API를 자연어로 조작할 수 있다.

설정 예시 (.mcp.json):
  {
    "mcpServers": {
      "synapse": {
        "command": "synapse",
        "args": ["mcp"]
      }
    }
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			return mcpserver.Serve(cmd.Context(), cfg)
		},
	}
}
