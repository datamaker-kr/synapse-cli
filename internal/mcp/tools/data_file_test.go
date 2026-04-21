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

func TestRegisterDataFile_NoPanic(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]config.ContextConfig{
			"test": {Server: "http://localhost"},
		},
	}
	RegisterDataFile(s, cfg)
}

func TestPresignedUpload_PostsCorrectPayload(t *testing.T) {
	var capturedPath string
	var capturedMethod string
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data":{"url":"https://s3.example.com/upload/abc"}}`)
	})
	defer ts.Close()

	payload := map[string]any{
		"data_unit":          1,
		"file_specification": 2,
		"file_name":          "car_001.jpg",
	}
	r, _, err := doPostRaw(context.Background(), sc, "/v2/data-files/presigned-upload/", payload)
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Equal(t, "POST", capturedMethod)
	assert.Equal(t, "/v2/data-files/presigned-upload/", capturedPath)
	assert.Equal(t, float64(1), capturedPayload["data_unit"])
	assert.Equal(t, float64(2), capturedPayload["file_specification"])
	assert.Equal(t, "car_001.jpg", capturedPayload["file_name"])
	assert.Contains(t, textFromResult(r), "s3.example.com")
}

func TestConfirmUpload_PostsCorrectPayload(t *testing.T) {
	var capturedPath string
	var capturedPayload map[string]any
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data":{"confirmed":true}}`)
	})
	defer ts.Close()

	payload := map[string]any{
		"data_unit":          1,
		"file_specification": 2,
	}
	r, _, err := doPostRaw(context.Background(), sc, "/v2/data-files/confirm-upload/", payload)
	require.NoError(t, err)
	assert.False(t, r.IsError)
	assert.Equal(t, "/v2/data-files/confirm-upload/", capturedPath)
	assert.Equal(t, float64(1), capturedPayload["data_unit"])
	// file_name은 confirm 단계에 포함되지 않음
	_, hasFileName := capturedPayload["file_name"]
	assert.False(t, hasFileName)
}

func TestPresignedUpload_ServerError(t *testing.T) {
	sc, ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":{"code":"NOT_FOUND","message":"data_unit not found"}}`)
	})
	defer ts.Close()

	r, _, err := doPostRaw(context.Background(), sc, "/v2/data-files/presigned-upload/", map[string]any{"data_unit": 999})
	require.NoError(t, err)
	assert.True(t, r.IsError)
}
