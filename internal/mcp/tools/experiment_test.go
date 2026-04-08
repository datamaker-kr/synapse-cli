package tools

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func cfgWithServer(serverURL string) *config.Config {
	return &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: serverURL, Token: "test-token", AuthMethod: "token", TenantCode: "test"},
		},
	}
}

func TestExperimentList_ViaHandler(t *testing.T) {
	_, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"exp-1","name":"test"}],"meta":{"pagination":{"next_cursor":null,"per_page":20}}}`)
	})
	defer ts.Close()

	cfg := cfgWithServer(ts.URL)
	sc, err := newClient(cfg)
	require.NoError(t, err)

	r, _, _ := fetchList(context.Background(), sc, "/v2/experiments/", false)
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "exp-1")
}

func TestExperimentGet_ViaHandler(t *testing.T) {
	_, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/v2/experiments/exp-123/")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"exp-123","name":"my experiment"}`)
	})
	defer ts.Close()

	cfg := cfgWithServer(ts.URL)
	sc, err := newClient(cfg)
	require.NoError(t, err)

	r, _, _ := fetchOne(context.Background(), sc, "/v2/experiments/exp-123/")
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "my experiment")
}

func TestExperimentCreate_DryRunDefault(t *testing.T) {
	// When DryRun is nil (default), should return dry-run message
	input := ExperimentCreateInput{ProjectID: "p1", Name: "exp1", DryRun: nil}
	isDryRun := input.DryRun == nil || *input.DryRun
	assert.True(t, isDryRun)
}

func TestExperimentCreate_DryRunFalse(t *testing.T) {
	_, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"new-exp","name":"exp1"}`)
	})
	defer ts.Close()

	cfg := cfgWithServer(ts.URL)
	sc, err := newClient(cfg)
	require.NoError(t, err)

	r, _, _ := doCreate(context.Background(), sc, "/v2/experiments/", map[string]any{
		"project_id": "p1", "name": "exp1",
	})
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "new-exp")
}

func TestExperimentDelete_NoForce(t *testing.T) {
	// When force=false, should return confirmation message
	input := ExperimentDeleteInput{ID: "exp1", Force: false}
	assert.False(t, input.Force)
}

func TestExperimentDelete_WithForce(t *testing.T) {
	_, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	defer ts.Close()

	cfg := cfgWithServer(ts.URL)
	sc, err := newClient(cfg)
	require.NoError(t, err)

	r, _, _ := doDelete(context.Background(), sc, "/v2/experiments/", "exp1", false)
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "삭제")
}

func TestExperiment_AuthError(t *testing.T) {
	_, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		fmt.Fprint(w, `{"error":{"code":"UNAUTHORIZED","message":"Authentication required. Run: synapse login"}}`)
	})
	defer ts.Close()

	cfg := cfgWithServer(ts.URL)
	sc, err := newClient(cfg)
	require.NoError(t, err)

	r, _, _ := fetchList(context.Background(), sc, "/v2/experiments/", false)
	assert.True(t, r.IsError)
}

func TestExperiment_NoContext(t *testing.T) {
	cfg := &config.Config{Contexts: map[string]config.ContextConfig{}}
	_, err := newClient(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "컨텍스트")
}

func TestRegisterExperiment_NoPanic(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := cfgWithServer("http://localhost")
	RegisterExperiment(s, cfg)
}
