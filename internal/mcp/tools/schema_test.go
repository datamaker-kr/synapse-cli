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

func TestRegisterSchema_NoPanic(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}
	RegisterSchema(s, cfg)
}

func TestSchemaFileSpecifications_NoCategory(t *testing.T) {
	var capturedPath string
	var capturedQuery string
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data":{"categories":{"image":{"file_specifications":{}}}}}`)
	})
	defer ts.Close()

	r, _, err := fetchOne(context.Background(), sc, "/v2/schemas/file-specifications/")
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Equal(t, "/v2/schemas/file-specifications/", capturedPath)
	assert.Empty(t, capturedQuery)
	assert.Contains(t, textFromResult(r), "categories")
}

func TestSchemaFileSpecifications_WithCategory(t *testing.T) {
	var capturedQuery string
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data":{"categories":{"image":{}}}}`)
	})
	defer ts.Close()

	path := addQueryParam("/v2/schemas/file-specifications/", "category", "image")
	r, _, err := fetchOne(context.Background(), sc, path)
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Contains(t, capturedQuery, "category=image")
}

func TestSchemaAnnotationConfigurations_WithCategory(t *testing.T) {
	var capturedPath string
	var capturedQuery string
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data":{"categories":{"image":{"annotation_types":[]}}}}`)
	})
	defer ts.Close()

	path := addQueryParam("/v2/schemas/annotation-configurations/", "category", "image")
	r, _, err := fetchOne(context.Background(), sc, path)
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Equal(t, "/v2/schemas/annotation-configurations/", capturedPath)
	assert.Contains(t, capturedQuery, "category=image")
}

func TestSchema_404OnUnsupportedBackend(t *testing.T) {
	// v2026.1.5 미만 백엔드는 404 반환
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":{"code":"NOT_FOUND","message":"endpoint not available"}}`)
	})
	defer ts.Close()

	r, _, err := fetchOne(context.Background(), sc, "/v2/schemas/file-specifications/")
	require.NoError(t, err)
	assert.True(t, r.IsError)
}
