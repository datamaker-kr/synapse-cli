package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

// textFromResult extracts text from the first Content element.
func textFromResult(r *mcp.CallToolResult) string {
	if len(r.Content) == 0 {
		return ""
	}
	if tc, ok := r.Content[0].(*mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

func TestToJSON(t *testing.T) {
	result := toJSON(map[string]string{"key": "value"})
	assert.Contains(t, result, `"key": "value"`)
}

func TestToolError(t *testing.T) {
	r, out, err := toolError("something failed")
	require.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, r.IsError)
	assert.Len(t, r.Content, 1)
	assert.Contains(t, textFromResult(r), "something failed")
}

func TestToolText(t *testing.T) {
	r, out, err := toolText("hello")
	require.NoError(t, err)
	assert.Nil(t, out)
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "hello")
}

func newTestServer(handler http.HandlerFunc) (*client.SynapseClient, *httptest.Server) {
	ts := httptest.NewServer(handler)
	cfg := &config.ContextConfig{
		Server:     ts.URL,
		Token:      "test-token",
		AuthMethod: "token",
		TenantCode: "test-tenant",
	}
	sc, _ := client.NewSynapseClient(cfg, "en")
	return sc, ts
}

func TestFetchOne_Success(t *testing.T) {
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"123","name":"test"}`)
	})
	defer ts.Close()

	r, _, err := fetchOne(context.Background(), sc, "/v2/test/123/")
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), `"id":"123"`)
}

func TestFetchOne_Error(t *testing.T) {
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":{"code":"NOT_FOUND","message":"Resource not found."}}`)
	})
	defer ts.Close()

	r, _, err := fetchOne(context.Background(), sc, "/v2/test/999/")
	require.NoError(t, err)
	assert.True(t, r.IsError)
}

func TestFetchList_SinglePage(t *testing.T) {
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"1"},{"id":"2"}],"meta":{"pagination":{"next_cursor":null,"per_page":20}}}`)
	})
	defer ts.Close()

	r, _, err := fetchList(context.Background(), sc, "/v2/test/", false)
	require.NoError(t, err)
	assert.False(t, r.IsError)
}

func TestFetchAllPages(t *testing.T) {
	callCount := 0
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			fmt.Fprint(w, `{"data":[{"id":"1"}],"meta":{"pagination":{"next_cursor":"abc","per_page":1}}}`)
		} else {
			fmt.Fprint(w, `{"data":[{"id":"2"}],"meta":{"pagination":{"next_cursor":null,"per_page":1}}}`)
		}
	})
	defer ts.Close()

	items, err := fetchAllPages(context.Background(), sc, "/v2/test/")
	require.NoError(t, err)
	assert.Len(t, items, 2)

	var item1, item2 map[string]string
	require.NoError(t, json.Unmarshal(items[0], &item1))
	require.NoError(t, json.Unmarshal(items[1], &item2))
	assert.Equal(t, "1", item1["id"])
	assert.Equal(t, "2", item2["id"])
}

func TestAddQueryParam(t *testing.T) {
	assert.Equal(t, "/v2/test/?cursor=abc", addQueryParam("/v2/test/", "cursor", "abc"))
	assert.Equal(t, "/v2/test/?status=running&cursor=abc", addQueryParam("/v2/test/?status=running", "cursor", "abc"))
}

func TestDoCreate_Success(t *testing.T) {
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"new-123","name":"created"}`)
	})
	defer ts.Close()

	r, _, err := doCreate(context.Background(), sc, "/v2/projects/", map[string]any{"name": "test"})
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "new-123")
}

func TestDoDelete_Success(t *testing.T) {
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	defer ts.Close()

	r, _, err := doDelete(context.Background(), sc, "/v2/projects/", "123")
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Contains(t, textFromResult(r), "삭제")
}
