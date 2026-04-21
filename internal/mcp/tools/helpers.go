package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

// newClient creates a SynapseClient from current config.
// Called per tool invocation to reflect latest ActiveContext (including env overrides).
func newClient(cfg *config.Config) (*client.SynapseClient, error) {
	ctxCfg, err := cfg.ActiveContext()
	if err != nil {
		return nil, fmt.Errorf("Synapse 컨텍스트가 설정되지 않았습니다. 'synapse config add-context'로 설정하세요: %w", err)
	}
	return client.NewSynapseClient(ctxCfg, "en")
}

// toJSON marshals v to a pretty-printed JSON string.
func toJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// toolError returns an MCP tool error result.
func toolError(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}, nil, nil
}

// toolText returns an MCP tool success result with text content.
func toolText(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

// fetchOne performs a GET request and returns the raw JSON body as a tool result.
func fetchOne(ctx context.Context, sc *client.SynapseClient, path string) (*mcp.CallToolResult, any, error) {
	resp, err := sc.RawRequest(ctx, "GET", path, nil)
	if err != nil {
		return toolError(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		apiErr := client.ParseAPIError(resp)
		return toolError(apiErr.FormatHuman())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return toolError(fmt.Sprintf("응답 읽기 실패: %v", err))
	}
	return toolText(string(body))
}

// fetchList performs a GET request with optional page-all support.
func fetchList(ctx context.Context, sc *client.SynapseClient, basePath string, pageAll bool) (*mcp.CallToolResult, any, error) {
	if pageAll {
		allData, err := fetchAllPages(ctx, sc, basePath)
		if err != nil {
			return toolError(err.Error())
		}
		return toolText(toJSON(allData))
	}

	resp, err := sc.RawRequest(ctx, "GET", basePath, nil)
	if err != nil {
		return toolError(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		apiErr := client.ParseAPIError(resp)
		return toolError(apiErr.FormatHuman())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return toolError(fmt.Sprintf("응답 읽기 실패: %v", err))
	}
	return toolText(string(body))
}

// fetchAllPages collects all pages via cursor-based pagination.
func fetchAllPages(ctx context.Context, sc *client.SynapseClient, basePath string) ([]json.RawMessage, error) {
	var all []json.RawMessage
	path := basePath

	for {
		resp, err := sc.RawRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		pr, err := client.ParsePaginatedResponse(body)
		if err != nil {
			return nil, err
		}

		var items []json.RawMessage
		if err := json.Unmarshal(pr.Data, &items); err != nil {
			return nil, fmt.Errorf("parse items: %w", err)
		}
		all = append(all, items...)

		if !pr.HasNextPage() {
			break
		}
		path = addQueryParam(basePath, "cursor", pr.NextCursor())
	}

	return all, nil
}

// buildListPath appends sort and fields query parameters to a base path.
func buildListPath(basePath, sort, fields string) string {
	path := basePath
	if sort != "" {
		path = addQueryParam(path, "sort", sort)
	}
	if fields != "" {
		path = addQueryParam(path, "fields", fields)
	}
	return path
}

// doCreateWithDryRun performs a POST with optional ?dry_run=true query parameter.
// Server-side dry-run validates payload without persisting (per v2 API policy).
func doCreateWithDryRun(ctx context.Context, sc *client.SynapseClient, path string, payload map[string]any, dryRun bool) (*mcp.CallToolResult, any, error) {
	if dryRun {
		path = addQueryParam(path, "dry_run", "true")
	}
	return doCreate(ctx, sc, path, payload)
}

// doPostRaw performs a POST to an endpoint that may not be registered in the OpenAPI schema
// (e.g., presigned-upload, confirm-upload, generate-tasks). Same behavior as doCreate.
func doPostRaw(ctx context.Context, sc *client.SynapseClient, path string, payload map[string]any) (*mcp.CallToolResult, any, error) {
	return doCreate(ctx, sc, path, payload)
}

// doCreate performs a POST request with a JSON body and returns the result.
func doCreate(ctx context.Context, sc *client.SynapseClient, path string, payload map[string]any) (*mcp.CallToolResult, any, error) {
	body, _ := json.Marshal(payload)
	resp, err := sc.RawRequest(ctx, "POST", path, strings.NewReader(string(body)))
	if err != nil {
		r, _, _ := toolError(err.Error())
		return r, nil, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		r, _, _ := toolError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
		return r, nil, nil
	}
	r, _, _ := toolText(string(respBody))
	return r, nil, nil
}

// doDelete performs a DELETE request and returns the result.
// If dryRun is true, appends ?dry_run=true to the request (permission check only).
func doDelete(ctx context.Context, sc *client.SynapseClient, basePath, id string, dryRun bool) (*mcp.CallToolResult, any, error) {
	path := basePath + id + "/"
	if dryRun {
		path += "?dry_run=true"
	}
	resp, err := sc.RawRequest(ctx, "DELETE", path, nil)
	if err != nil {
		r, _, _ := toolError(err.Error())
		return r, nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		r, _, _ := toolError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
		return r, nil, nil
	}
	r, _, _ := toolText(fmt.Sprintf("리소스 '%s' 가 삭제되었습니다.", id))
	return r, nil, nil
}

// addQueryParam appends a query parameter to a URL path.
func addQueryParam(basePath, key, value string) string {
	sep := "?"
	if strings.Contains(basePath, "?") {
		sep = "&"
	}
	return basePath + sep + key + "=" + url.QueryEscape(value)
}
