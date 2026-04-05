package resources

import (
	"context"
	_ "embed"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed default_skill.md
var defaultSkill string

func loadSkillContent() string {
	if content, err := os.ReadFile("skills/synapse-cli.md"); err == nil {
		return string(content)
	}
	return defaultSkill
}

// Register adds MCP resources to the server.
func Register(s *mcp.Server) {
	content := loadSkillContent()
	s.AddResource(&mcp.Resource{
		URI:      "synapse://skills/synapse-cli",
		Name:     "Synapse CLI Skill Guide",
		MIMEType: "text/markdown",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "synapse://skills/synapse-cli", Text: content, MIMEType: "text/markdown"},
			},
		}, nil
	})
}
