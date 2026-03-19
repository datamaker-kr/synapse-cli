package branding

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintLogo_NoColor(t *testing.T) {
	buf := new(bytes.Buffer)
	PrintLogo(buf, "1.0.0", false)

	out := buf.String()
	assert.Contains(t, out, "CLI", "logo should contain CLI text")
	assert.Contains(t, out, "1.0.0", "logo should contain version")
	assert.NotContains(t, out, "\033[", "no color should have no ANSI escape codes")
}

func TestPrintLogo_WithColor(t *testing.T) {
	buf := new(bytes.Buffer)
	PrintLogo(buf, "dev", true)

	out := buf.String()
	assert.Contains(t, out, "\033[", "color mode should have ANSI escape codes")
	assert.Contains(t, out, "dev", "logo should contain version")
}

func TestPrintLogo_VersionInterpolation(t *testing.T) {
	buf := new(bytes.Buffer)
	PrintLogo(buf, "v2.3.4", false)

	out := buf.String()
	assert.Contains(t, out, "v2.3.4")
	// Should not contain the raw format directive
	assert.NotContains(t, out, "%s")
}

func TestPrintLogo_AsciiArtStructure(t *testing.T) {
	buf := new(bytes.Buffer)
	PrintLogo(buf, "test", false)

	lines := strings.Split(buf.String(), "\n")
	// Logo should have multiple lines
	assert.Greater(t, len(lines), 10, "logo should have at least 10 lines")
}
