package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type AssignmentListInput struct {
	TaskID  string `json:"task_id,omitempty" jsonschema:"태스크 ID로 필터링"`
	Sort    string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields  string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name). context window 최적화용"`
	PageAll bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

// RegisterAssignment registers assignment-related MCP tools.
func RegisterAssignment(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_assignment_list",
		Description: "Synapse 할당 목록을 조회한다. 기본 per_page=50, 최대 200. 태스크 ID로 필터링 가능.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input AssignmentListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/assignments/"
			if input.TaskID != "" {
				path += "?task=" + url.QueryEscape(input.TaskID)
			}
			path = buildListPath(path, input.Sort, input.Fields)
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})
}
