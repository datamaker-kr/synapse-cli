package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func TestRegisterProject_NoPanic(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}
	RegisterProject(s, cfg)
}

func TestProjectCreate_ConfigurationParsing(t *testing.T) {
	// 정상 configuration JSON 파싱
	validConfig := `{"schema_type":"dm_schema","classification":{"bounding_box":{"id":"abc","representativeCodes":[],"classification_schema":[]}}}`
	var configuration any
	err := json.Unmarshal([]byte(validConfig), &configuration)
	require.NoError(t, err)

	configMap := configuration.(map[string]any)
	assert.Equal(t, "dm_schema", configMap["schema_type"])
	assert.Contains(t, configMap, "classification")
}

func TestProjectCreate_EmptyConfiguration(t *testing.T) {
	// 빈 configuration("{}")도 정상 파싱
	emptyConfig := `{}`
	var configuration any
	err := json.Unmarshal([]byte(emptyConfig), &configuration)
	require.NoError(t, err)

	configMap := configuration.(map[string]any)
	assert.Empty(t, configMap)
}

func TestProjectCreate_InvalidConfigurationJSON(t *testing.T) {
	invalidConfig := `{schema_type: dm_schema}` // 잘못된 JSON (unquoted key)
	var configuration any
	err := json.Unmarshal([]byte(invalidConfig), &configuration)
	assert.Error(t, err)
}

func TestProjectCreate_DataCollectionNullable(t *testing.T) {
	// data_collection이 nil이면 payload에서 제외
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"data":{"id":1}}`)
	})
	defer ts.Close()

	payload := map[string]any{
		"title":         "test",
		"category":      "image",
		"configuration": map[string]any{},
	}
	_, _, err := doCreate(context.Background(), sc, "/v2/projects/", payload)
	require.NoError(t, err)
	_, hasDataCollection := capturedPayload["data_collection"]
	assert.False(t, hasDataCollection)
}

func TestProjectCreate_DataCollectionIncluded(t *testing.T) {
	// data_collection이 있으면 payload에 포함
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"data":{"id":1}}`)
	})
	defer ts.Close()

	dcID := 42
	payload := map[string]any{
		"title":           "test",
		"category":        "image",
		"configuration":   map[string]any{},
		"data_collection": dcID,
	}
	_, _, err := doCreate(context.Background(), sc, "/v2/projects/", payload)
	require.NoError(t, err)
	assert.Equal(t, float64(42), capturedPayload["data_collection"]) // JSON unmarshal int → float64
}

func TestProjectCreate_TitleRequired(t *testing.T) {
	// title 필드가 payload에 포함되는지 검증
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"data":{"id":1}}`)
	})
	defer ts.Close()

	payload := map[string]any{
		"title":         "Vehicle Detection",
		"category":      "image",
		"configuration": map[string]any{},
	}
	_, _, err := doCreate(context.Background(), sc, "/v2/projects/", payload)
	require.NoError(t, err)
	assert.Equal(t, "Vehicle Detection", capturedPayload["title"])
	// "name" 필드는 사용하지 않음 (API 스키마 정합성)
	_, hasName := capturedPayload["name"]
	assert.False(t, hasName)
}

func TestProjectGenerateTasks_DryRunPath(t *testing.T) {
	var capturedPath string
	var capturedQuery string
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(202)
		fmt.Fprint(w, `{"data":{"job_id":"job_x","status":"queued"}}`)
	})
	defer ts.Close()

	// dry-run 모드에서 ?dry_run=true 쿼리 추가 확인
	path := "/v2/projects/123/generate-tasks/"
	path = addQueryParam(path, "dry_run", "true")
	r, _, err := doPostRaw(context.Background(), sc, path, map[string]any{})
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Equal(t, "/v2/projects/123/generate-tasks/", capturedPath)
	assert.Contains(t, capturedQuery, "dry_run=true")
}

func TestProjectCreate_TimeSeriesCategory(t *testing.T) {
	// time_series 카테고리 지원 검증 (서버에 그대로 전달)
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"data":{"id":1}}`)
	})
	defer ts.Close()

	payload := map[string]any{
		"title":         "ts-project",
		"category":      "time_series",
		"configuration": map[string]any{},
	}
	_, _, err := doCreate(context.Background(), sc, "/v2/projects/", payload)
	require.NoError(t, err)
	assert.Equal(t, "time_series", capturedPayload["category"])
}
