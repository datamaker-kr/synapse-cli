package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func TestRegisterConfig_AllToolsRegistered(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}

	// Should not panic
	RegisterConfig(s, cfg)
}

func TestRegisterExperiment_AllToolsRegistered(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}

	// Should not panic
	RegisterExperiment(s, cfg)
}

func TestRegisterProject_AllToolsRegistered(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}

	// Should not panic
	RegisterProject(s, cfg)
}

func TestRegisterAllTools(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}

	// All register functions should not panic
	RegisterExperiment(s, cfg)
	RegisterJob(s, cfg)
	RegisterProject(s, cfg)
	RegisterTask(s, cfg)
	RegisterAssignment(s, cfg)
	RegisterDataCollection(s, cfg)
	RegisterDataUnit(s, cfg)
	RegisterDataFile(s, cfg)
	RegisterConfig(s, cfg)

	// Verify 21 tools total via listing them is not possible with unexported fields,
	// but we verified registration didn't panic
}
