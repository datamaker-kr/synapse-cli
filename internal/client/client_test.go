package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func TestNewSynapseClient_TokenAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Token test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Token test-token")
		}
		if got := r.Header.Get("SYNAPSE-Tenant"); got != "ws-001" {
			t.Errorf("SYNAPSE-Tenant = %q, want %q", got, "ws-001")
		}
		if got := r.Header.Get("Accept-Language"); got != "ko" {
			t.Errorf("Accept-Language = %q, want %q", got, "ko")
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))
	defer server.Close()

	cfg := &config.ContextConfig{
		Server:     server.URL,
		AuthMethod: "token",
		Token:      "test-token",
		TenantCode: "ws-001",
	}

	client, err := NewSynapseClient(cfg, "ko")
	if err != nil {
		t.Fatalf("NewSynapseClient failed: %v", err)
	}

	resp, err := client.RawRequest(context.Background(), "GET", "/v2/tenants/", nil)
	if err != nil {
		t.Fatalf("RawRequest failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNewSynapseClient_AccessTokenAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("SYNAPSE-ACCESS-TOKEN"); got != "syn_test123" {
			t.Errorf("SYNAPSE-ACCESS-TOKEN = %q, want %q", got, "syn_test123")
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization should be empty for access_token, got %q", got)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &config.ContextConfig{
		Server:      server.URL,
		AuthMethod:  "access_token",
		AccessToken: "syn_test123",
	}

	client, err := NewSynapseClient(cfg, "en")
	if err != nil {
		t.Fatalf("NewSynapseClient failed: %v", err)
	}

	resp, err := client.RawRequest(context.Background(), "GET", "/v2/projects/", nil)
	if err != nil {
		t.Fatalf("RawRequest failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRawRequest_PostWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.WriteHeader(201)
	}))
	defer server.Close()

	cfg := &config.ContextConfig{Server: server.URL, Token: "tok"}
	client, _ := NewSynapseClient(cfg, "en")

	resp, err := client.RawRequest(context.Background(), "POST", "/v2/projects/", strings.NewReader(`{"name":"test"}`))
	if err != nil {
		t.Fatalf("RawRequest failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestHealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health/" {
			t.Errorf("expected /health/, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	if err := HealthCheck(context.Background(), server.URL); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestHealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer server.Close()

	if err := HealthCheck(context.Background(), server.URL); err == nil {
		t.Fatal("expected error for 503")
	}
}
