package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func TestHealthCommand_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health/", r.URL.Path)
		w.WriteHeader(200)
	}))
	defer server.Close()

	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)
	t.Setenv("SYNAPSE_NO_LOGO", "1")

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts:       map[string]config.ContextConfig{"test": {Server: server.URL}},
	}
	require.NoError(t, cfg.Save())

	stdout, _, err := executeCommand("health")
	require.NoError(t, err)
	assert.Contains(t, stdout, "OK")
	assert.Contains(t, stdout, server.URL)
}

func TestHealthCommand_NoContext(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)
	t.Setenv("SYNAPSE_NO_LOGO", "1")

	_, _, err := executeCommand("health")
	assert.Error(t, err)
}
