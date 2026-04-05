package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type ExperimentListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"실험 상태 필터 (running|completed|failed 등)"`
	PageAll bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type ExperimentGetInput struct {
	ID string `json:"id" jsonschema:"실험 ID"`
}

type ExperimentCreateInput struct {
	ProjectID   string `json:"project_id" jsonschema:"프로젝트 ID"`
	Name        string `json:"name" jsonschema:"실험 이름"`
	Description string `json:"description,omitempty" jsonschema:"실험 설명"`
	DryRun      *bool  `json:"dry_run,omitempty" jsonschema:"true이면 실제 생성하지 않고 요청 내용만 반환. 기본값 true"`
}

type ExperimentDeleteInput struct {
	ID    string `json:"id" jsonschema:"삭제할 실험 ID"`
	Force bool   `json:"force" jsonschema:"true로 설정해야 실제 삭제 실행"`
}

// RegisterExperiment registers experiment-related MCP tools.
func RegisterExperiment(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_experiment_list",
		Description: "Synapse 실험 목록을 조회한다. 상태별 필터링 가능.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ExperimentListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/experiments/"
			if input.Status != "" {
				path += "?status=" + url.QueryEscape(input.Status)
			}
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_experiment_get",
		Description: "Synapse 실험 상세 정보를 조회한다.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ExperimentGetInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchOne(ctx, sc, "/v2/experiments/"+input.ID+"/")
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_experiment_create",
		Description: "Synapse 실험을 생성한다. dry_run 기본 활성화.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ExperimentCreateInput) (*mcp.CallToolResult, any, error) {
			isDryRun := input.DryRun == nil || *input.DryRun
			if isDryRun {
				payload := map[string]any{"project_id": input.ProjectID, "name": input.Name, "description": input.Description}
				r, _, _ := toolText("[DRY RUN] POST /v2/experiments/ 로 다음 데이터를 전송합니다:\n" + toJSON(payload) + "\n\n실행하려면 dry_run=false로 다시 호출하세요.")
				return r, nil, nil
			}
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doCreate(ctx, sc, "/v2/experiments/", map[string]any{
				"project_id": input.ProjectID, "name": input.Name, "description": input.Description,
			})
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_experiment_delete",
		Description: "Synapse 실험을 삭제한다. force=true 필수.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ExperimentDeleteInput) (*mcp.CallToolResult, any, error) {
			if !input.Force {
				r, _, _ := toolText("실험 '" + input.ID + "' 삭제를 요청했습니다. 실제 삭제하려면 force=true로 다시 호출하세요.")
				return r, nil, nil
			}
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doDelete(ctx, sc, "/v2/experiments/", input.ID)
		})
}
