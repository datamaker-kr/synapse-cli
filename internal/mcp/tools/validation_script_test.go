package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func TestRegisterValidationScript_NoPanic(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}
	RegisterValidationScript(s, cfg)
}
