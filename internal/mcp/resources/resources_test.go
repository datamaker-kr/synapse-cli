package resources

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestLoadSkillContent(t *testing.T) {
	content := loadSkillContent()
	// Should return embedded default content (skills/ file won't exist in test context)
	assert.Contains(t, content, "Synapse")
	assert.Contains(t, content, "플랫폼")
}

func TestRegister(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	// Should not panic
	Register(s)
}
