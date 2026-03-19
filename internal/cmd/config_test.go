package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func TestConfigListContexts_Empty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)

	stdout, _, err := executeCommand("config", "list-contexts")
	require.NoError(t, err)
	assert.Contains(t, stdout, "No contexts configured")
}

func TestConfigCurrentContext_NoContext(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)

	_, _, err := executeCommand("config", "current-context")
	assert.Error(t, err)
}

func TestConfigView_WithContext(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)
	t.Setenv("SYNAPSE_NO_LOGO", "1")

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {
				Server:     "https://test.example.com",
				AuthMethod: "token",
				Token:      "secret-token-12345678",
				TenantCode: "ws-test",
			},
		},
	}
	require.NoError(t, cfg.Save())

	stdout, _, err := executeCommand("config", "view")
	require.NoError(t, err)
	assert.Contains(t, stdout, "test.example.com")
	assert.Contains(t, stdout, "ws-test")
	// Token should be masked
	assert.NotContains(t, stdout, "secret-token-12345678")
	assert.Contains(t, stdout, "***...")
}

func TestConfigSetLanguage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts:       map[string]config.ContextConfig{"test": {Server: "https://test.com"}},
	}
	require.NoError(t, cfg.Save())

	_, _, err := executeCommand("config", "set-language", "ko")
	require.NoError(t, err)

	loaded, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "ko", loaded.Language)
}

func TestConfigSetLanguage_Invalid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)

	_, _, err := executeCommand("config", "set-language", "fr")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unsupported"))
}

func TestConfigListContexts_WithContexts(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)
	t.Setenv("SYNAPSE_NO_LOGO", "1")

	cfg := &config.Config{
		CurrentContext: "prod",
		Contexts: map[string]config.ContextConfig{
			"prod":    {Server: "https://prod.example.com"},
			"staging": {Server: "https://staging.example.com"},
		},
	}
	require.NoError(t, cfg.Save())

	stdout, _, err := executeCommand("config", "list-contexts")
	require.NoError(t, err)
	assert.Contains(t, stdout, "prod")
	assert.Contains(t, stdout, "staging")
	assert.Contains(t, stdout, "* ") // active context marker
}
