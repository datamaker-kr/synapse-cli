package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type SchemaFileSpecInput struct {
	Category string `json:"category,omitempty" jsonschema:"카테고리 필터 (image|video|audio|text|pcd|data). 미지정 시 전체 반환"`
}

type SchemaAnnotationConfigInput struct {
	Category string `json:"category,omitempty" jsonschema:"카테고리 필터 (image|video|audio|text|pcd|prompt|time_series). 미지정 시 전체 반환"`
}

// RegisterSchema registers Schema Discovery MCP tools.
// Requires Synapse Backend v2026.1.5+.
func RegisterSchema(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "synapse_schema_file_specifications",
		Description: "data-collection 생성을 위한 file_specifications 스키마 메타데이터를 조회한다. " +
			"validation_rules (naming 패턴: {spec_key}_{index}, primary 1개 필수, function_type=main 1개 필수), " +
			"카테고리별 지원 파일 타입, payload 예시 반환. " +
			"data-collection 생성 전 반드시 조회할 것. " +
			"요구 사항: Synapse Backend v2026.1.5+",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input SchemaFileSpecInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/schemas/file-specifications/"
			if input.Category != "" {
				path = addQueryParam(path, "category", url.QueryEscape(input.Category))
			}
			r, _, _ := fetchOne(ctx, sc, path)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name: "synapse_schema_annotation_configurations",
		Description: "project 생성을 위한 configuration JSON 스키마를 조회한다. " +
			"카테고리별 annotation 타입, classification 구조, widget 타입(select/radio/multi_select/text), " +
			"UUID v4 생성 규칙 반환. " +
			"project 생성 전 반드시 조회할 것. " +
			"요구 사항: Synapse Backend v2026.1.5+",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input SchemaAnnotationConfigInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/schemas/annotation-configurations/"
			if input.Category != "" {
				path = addQueryParam(path, "category", url.QueryEscape(input.Category))
			}
			r, _, _ := fetchOne(ctx, sc, path)
			return r, nil, nil
		})
}
