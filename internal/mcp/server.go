package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
	"github.com/datamaker-kr/synapse-cli/internal/mcp/resources"
	"github.com/datamaker-kr/synapse-cli/internal/mcp/tools"
)

var version = "dev"

// SetVersion sets the MCP server version (called from main).
func SetVersion(v string) {
	version = v
}

// Serve starts the MCP server on stdio transport.
func Serve(ctx context.Context, cfg *config.Config) error {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "synapse",
			Version: version,
		},
		&mcp.ServerOptions{
			Instructions: "Synapse ML 플랫폼 API 도구. 실험/잡/프로젝트 등 리소스를 조회하고 관리한다.",
		},
	)

	// Register MCP resources
	resources.Register(s)

	// Register MCP tools
	tools.RegisterExperiment(s, cfg)
	tools.RegisterJob(s, cfg)
	tools.RegisterProject(s, cfg)
	tools.RegisterTask(s, cfg)
	tools.RegisterAssignment(s, cfg)
	tools.RegisterDataCollection(s, cfg)
	tools.RegisterDataUnit(s, cfg)
	tools.RegisterDataFile(s, cfg)
	tools.RegisterConfig(s, cfg)
	tools.RegisterValidationScript(s, cfg)
	tools.RegisterSchema(s, cfg)

	return s.Run(ctx, &mcp.StdioTransport{})
}
