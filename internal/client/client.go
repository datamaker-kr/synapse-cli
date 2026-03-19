package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/datamaker-kr/synapse-cli/internal/client/generated"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

// SynapseClient wraps the generated client with auth injection and error handling.
type SynapseClient struct {
	Generated  *generated.Client
	Config     *config.ContextConfig
	HTTPClient *http.Client
	Lang       string
}

// NewSynapseClient creates a new client from context config.
func NewSynapseClient(cfg *config.ContextConfig, lang string) (*SynapseClient, error) {
	if cfg.Server == "" {
		return nil, fmt.Errorf("server URL is not configured")
	}

	sc := &SynapseClient{
		Config: cfg,
		Lang:   lang,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	server := strings.TrimRight(cfg.Server, "/")
	gen, err := generated.NewClient(server, generated.WithHTTPClient(sc.HTTPClient), generated.WithRequestEditorFn(sc.authEditor))
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	sc.Generated = gen

	return sc, nil
}

// authEditor injects authentication and language headers into every request.
func (c *SynapseClient) authEditor(ctx context.Context, req *http.Request) error {
	switch c.Config.AuthMethod {
	case "access_token":
		if c.Config.AccessToken != "" {
			req.Header.Set("SYNAPSE-ACCESS-TOKEN", c.Config.AccessToken)
		}
	default:
		if c.Config.Token != "" {
			req.Header.Set("Authorization", "Token "+c.Config.Token)
		}
		if c.Config.TenantCode != "" {
			req.Header.Set("SYNAPSE-Tenant", c.Config.TenantCode)
		}
	}

	// i18n: Accept-Language header
	if c.Lang != "" {
		req.Header.Set("Accept-Language", c.Lang)
	}

	return nil
}

// RawRequest performs an arbitrary HTTP request with auth headers injected.
// Used by the escape hatch `synapse api` command.
func (c *SynapseClient) RawRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := strings.TrimRight(c.Config.Server, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if err := c.authEditor(ctx, req); err != nil {
		return nil, fmt.Errorf("inject auth: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.HTTPClient.Do(req)
}

// HealthCheck performs a health check against the server.
func HealthCheck(ctx context.Context, serverURL string) error {
	url := strings.TrimRight(serverURL, "/") + "/health/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: HTTP %d", resp.StatusCode)
	}
	return nil
}
