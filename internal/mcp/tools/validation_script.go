package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type ValidationScriptListInput struct {
	IsActive *bool  `json:"is_active,omitempty" jsonschema:"활성 상태 필터"`
	Search   string `json:"search,omitempty" jsonschema:"이름/설명 검색"`
	Sort     string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields   string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name,is_active). context window 최적화용"`
	PageAll  bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지 조회. 기본 per_page=50, 최대 200"`
}

type ValidationScriptGetInput struct {
	ID string `json:"id" jsonschema:"검증 스크립트 ID"`
}

// RegisterValidationScript registers validation-script-related MCP tools.
func RegisterValidationScript(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_validation_script_list",
		Description: "Synapse 검증 스크립트 목록을 조회한다. is_active, search 필터 가능. 기본 per_page=50, 최대 200.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ValidationScriptListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/validation-scripts/"
			if input.IsActive != nil {
				path = addQueryParam(path, "is_active", fmt.Sprintf("%t", *input.IsActive))
			}
			if input.Search != "" {
				path = addQueryParam(path, "search", url.QueryEscape(input.Search))
			}
			path = buildListPath(path, input.Sort, input.Fields)
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_validation_script_get",
		Description: "Synapse 검증 스크립트 상세 정보를 조회한다. code, description, timeout 포함.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ValidationScriptGetInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchOne(ctx, sc, "/v2/validation-scripts/"+input.ID+"/")
			return r, nil, nil
		})
}
