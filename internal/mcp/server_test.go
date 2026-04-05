package mcpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetVersion(t *testing.T) {
	SetVersion("1.2.3")
	assert.Equal(t, "1.2.3", version)
	// Reset
	SetVersion("dev")
}
