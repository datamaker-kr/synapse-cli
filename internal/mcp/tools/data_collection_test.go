package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataCollectionCreate_DryRunDefault_AddsQueryParam(t *testing.T) {
	var capturedQuery string
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data":{"dry_run":true,"action":"create"}}`)
	})
	defer ts.Close()

	payload := map[string]any{"name": "test", "category": "image"}
	r, _, err := doCreateWithDryRun(context.Background(), sc, "/v2/data-collections/", payload, true)
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Contains(t, capturedQuery, "dry_run=true")
	assert.Equal(t, "test", capturedPayload["name"])
	assert.Equal(t, "image", capturedPayload["category"])
}

func TestDataCollectionCreate_FileSpecificationsParsing(t *testing.T) {
	// 정상 JSON 배열 파싱 검증
	validJSON := `[{"name":"image_1","file_type":"image","is_required":true,"is_primary":true,"function_type":"main","index":1}]`
	var fileSpecs []any
	err := json.Unmarshal([]byte(validJSON), &fileSpecs)
	require.NoError(t, err)
	require.Len(t, fileSpecs, 1)

	first := fileSpecs[0].(map[string]any)
	assert.Equal(t, "image_1", first["name"])
	assert.Equal(t, "main", first["function_type"])
	assert.Equal(t, true, first["is_primary"])
}

func TestDataCollectionCreate_InvalidFileSpecificationsJSON(t *testing.T) {
	// 잘못된 JSON은 파싱 에러
	invalidJSON := `not a json`
	var fileSpecs []any
	err := json.Unmarshal([]byte(invalidJSON), &fileSpecs)
	assert.Error(t, err)
}

func TestDataCollectionCreate_FullPayload(t *testing.T) {
	// payload에 file_specifications가 그대로 전달되는지 통합 검증
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(201)
		fmt.Fprint(w, `{"data":{"id":1}}`)
	})
	defer ts.Close()

	fileSpecs := []any{
		map[string]any{
			"name":          "image_1",
			"file_type":     "image",
			"is_required":   true,
			"is_primary":    true,
			"function_type": "main",
			"index":         1,
		},
	}
	payload := map[string]any{
		"name":                "vehicle-dataset",
		"category":            "image",
		"description":         "test",
		"file_specifications": fileSpecs,
	}

	_, _, err := doCreate(context.Background(), sc, "/v2/data-collections/", payload)
	require.NoError(t, err)

	specs, ok := capturedPayload["file_specifications"].([]any)
	require.True(t, ok)
	assert.Len(t, specs, 1)
}
