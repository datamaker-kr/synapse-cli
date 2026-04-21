package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type ProjectListInput struct {
	Sort    string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields  string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name). context window 최적화용"`
	PageAll bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type ProjectGetInput struct {
	ID string `json:"id" jsonschema:"프로젝트 ID"`
}

type ProjectCreateInput struct {
	Title          string `json:"title" jsonschema:"프로젝트 제목 (필수)"`
	Category       string `json:"category" jsonschema:"카테고리 (image|video|audio|text|pcd|prompt|time_series) (필수)"`
	Configuration  string `json:"configuration" jsonschema:"configuration JSON 문자열 (필수). schema_type + classification 구조. UUID v4 생성 필수. 빈 값은 '{}' 전달. 먼저 synapse_schema_annotation_configurations로 스키마 조회 권장"`
	DataCollection *int   `json:"data_collection,omitempty" jsonschema:"연결할 data-collection ID (optional, nullable)"`
	Description    string `json:"description,omitempty" jsonschema:"프로젝트 설명"`
	DryRun         *bool  `json:"dry_run,omitempty" jsonschema:"true이면 dry-run (서버 validation만 수행). 기본값 true"`
}

type ProjectGenerateTasksInput struct {
	ProjectID string `json:"project_id" jsonschema:"프로젝트 ID (필수)"`
	DryRun    *bool  `json:"dry_run,omitempty" jsonschema:"true이면 dry-run. 기본값 true"`
}

type ProjectDeleteInput struct {
	ID     string `json:"id" jsonschema:"삭제할 프로젝트 ID"`
	Force  bool   `json:"force" jsonschema:"true로 설정해야 삭제 진행"`
	DryRun *bool  `json:"dry_run,omitempty" jsonschema:"true이면 권한 체크만 수행. 기본값 true"`
}

// RegisterProject registers project-related MCP tools.
func RegisterProject(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_project_list",
		Description: "Synapse 프로젝트 목록을 조회한다. 기본 per_page=50, 최대 200.",
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
		Name: "synapse_project_create",
		Description: "Synapse 프로젝트를 생성한다. dry_run 기본 활성화. " +
			"먼저 synapse_schema_annotation_configurations로 카테고리별 configuration 스키마 조회 권장. " +
			"configuration은 JSON 문자열로 전달 (예: '{\"schema_type\":\"dm_schema\",\"classification\":{...}}', 빈 값은 '{}'). " +
			"UUID v4는 직접 생성하여 classification.id 등에 주입.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectCreateInput) (*mcp.CallToolResult, any, error) {
			var configuration any = map[string]any{}
			if input.Configuration != "" {
				if err := json.Unmarshal([]byte(input.Configuration), &configuration); err != nil {
					r, _, _ := toolError(fmt.Sprintf("configuration이 유효한 JSON이 아닙니다: %v", err))
					return r, nil, nil
				}
			}
			payload := map[string]any{
				"title":         input.Title,
				"category":      input.Category,
				"configuration": configuration,
			}
			if input.Description != "" {
				payload["description"] = input.Description
			}
			if input.DataCollection != nil {
				payload["data_collection"] = *input.DataCollection
			}

			isDryRun := input.DryRun == nil || *input.DryRun
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doCreateWithDryRun(ctx, sc, "/v2/projects/", payload, isDryRun)
		})

	mcp.AddTool(s, &mcp.Tool{
		Name: "synapse_project_generate_tasks",
		Description: "프로젝트에 task를 자동 생성한다. dry_run 기본 활성화. " +
			"전제 조건: data-unit의 can_generate_task=true 상태여야 함 (비동기 파일 처리 완료 후 자동 설정). " +
			"비동기 처리되어 202 Accepted + job_id 응답 가능.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectGenerateTasksInput) (*mcp.CallToolResult, any, error) {
			isDryRun := input.DryRun == nil || *input.DryRun
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/projects/" + input.ProjectID + "/generate-tasks/"
			if isDryRun {
				path = addQueryParam(path, "dry_run", "true")
			}
			return doPostRaw(ctx, sc, path, map[string]any{})
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_project_delete",
		Description: "Synapse 프로젝트를 삭제한다. force=true 필수. dry_run 기본 활성화.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input ProjectDeleteInput) (*mcp.CallToolResult, any, error) {
			if !input.Force {
				r, _, _ := toolText("프로젝트 '" + input.ID + "' 삭제를 요청했습니다. 실제 삭제하려면 force=true로 다시 호출하세요.")
				return r, nil, nil
			}
			isDryRun := input.DryRun == nil || *input.DryRun
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doDelete(ctx, sc, "/v2/projects/", input.ID, isDryRun)
		})
}
