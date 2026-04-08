package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type TaskListInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"프로젝트 ID로 필터링"`
	Sort      string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields    string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name). context window 최적화용"`
	PageAll   bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

// RegisterTask registers task-related MCP tools.
func RegisterTask(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_task_list",
		Description: "Synapse 태스크 목록을 조회한다. 기본 per_page=50, 최대 200. 프로젝트 ID로 필터링 가능.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input TaskListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/tasks/"
			if input.ProjectID != "" {
				path += "?project=" + url.QueryEscape(input.ProjectID)
			}
			path = buildListPath(path, input.Sort, input.Fields)
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})
}
