package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type ProjectListInput struct {
	PageAll bool `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type ProjectGetInput struct {
	ID string `json:"id" jsonschema:"프로젝트 ID"`
}

type ProjectCreateInput struct {
	Name        string `json:"name" jsonschema:"프로젝트 이름"`
	Description string `json:"description,omitempty" jsonschema:"프로젝트 설명"`
	DryRun      *bool  `json:"dry_run,omitempty" jsonschema:"true이면 실제 생성하지 않고 요청 내용만 반환. 기본값 true"`
}

type ProjectDeleteInput struct {
	ID    string `json:"id" jsonschema:"삭제할 프로젝트 ID"`
	Force bool   `json:"force" jsonschema:"true로 설정해야 실제 삭제 실행"`
}

// RegisterProject registers project-related MCP tools.
func RegisterProject(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_project_list",
		Description: "Synapse 프로젝트 목록을 조회한다.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchList(ctx, sc, "/v2/projects/", input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_project_get",
		Description: "Synapse 프로젝트 상세 정보를 조회한다.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectGetInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchOne(ctx, sc, "/v2/projects/"+input.ID+"/")
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_project_create",
		Description: "Synapse 프로젝트를 생성한다. dry_run 기본 활성화.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectCreateInput) (*mcp.CallToolResult, any, error) {
			isDryRun := input.DryRun == nil || *input.DryRun
			if isDryRun {
				payload := map[string]any{"name": input.Name, "description": input.Description}
				r, _, _ := toolText("[DRY RUN] POST /v2/projects/ 로 다음 데이터를 전송합니다:\n" + toJSON(payload) + "\n\n실행하려면 dry_run=false로 다시 호출하세요.")
				return r, nil, nil
			}
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doCreate(ctx, sc, "/v2/projects/", map[string]any{
				"name": input.Name, "description": input.Description,
			})
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_project_delete",
		Description: "Synapse 프로젝트를 삭제한다. force=true 필수.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectDeleteInput) (*mcp.CallToolResult, any, error) {
			if !input.Force {
				r, _, _ := toolText("프로젝트 '" + input.ID + "' 삭제를 요청했습니다. 실제 삭제하려면 force=true로 다시 호출하세요.")
				return r, nil, nil
			}
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doDelete(ctx, sc, "/v2/projects/", input.ID)
		})
}
