package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type DataCollectionListInput struct {
	Sort    string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields  string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name). context window 최적화용"`
	PageAll bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type DataCollectionGetInput struct {
	ID string `json:"id" jsonschema:"데이터 컬렉션 ID"`
}

type DataCollectionCreateInput struct {
	Name               string `json:"name" jsonschema:"데이터 컬렉션 이름 (필수)"`
	Category           string `json:"category" jsonschema:"카테고리 (image|video|audio|text|pcd|data) (필수)"`
	Description        string `json:"description,omitempty" jsonschema:"데이터 컬렉션 설명"`
	FileSpecifications string `json:"file_specifications,omitempty" jsonschema:"file_specifications JSON 배열 문자열. 예: [{\"name\":\"image_1\",\"file_type\":\"image\",\"is_required\":true,\"is_primary\":true,\"function_type\":\"main\",\"index\":1}]. naming은 {spec_key}_{index} 형식 필수. is_primary=true 1개 + function_type=main 1개 필수. 먼저 synapse_schema_file_specifications로 스키마 조회 권장"`
	DryRun             *bool  `json:"dry_run,omitempty" jsonschema:"true이면 dry-run (서버 validation만 수행). 기본값 true"`
}

// RegisterDataCollection registers data-collection-related MCP tools.
func RegisterDataCollection(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_data_collection_list",
		Description: "Synapse 데이터 컬렉션 목록을 조회한다. 기본 per_page=50, 최대 200.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataCollectionListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := buildListPath("/v2/data-collections/", input.Sort, input.Fields)
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_data_collection_get",
		Description: "Synapse 데이터 컬렉션 상세 정보를 조회한다. 결과에 file_specifications (유닛당 파일 구성 정의) 포함.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataCollectionGetInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchOne(ctx, sc, "/v2/data-collections/"+input.ID+"/")
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name: "synapse_data_collection_create",
		Description: "Synapse 데이터 컬렉션을 생성한다. dry_run 기본 활성화. " +
			"먼저 synapse_schema_file_specifications로 스키마 조회 후 file_specifications를 구성할 것. " +
			"file_specifications는 JSON 배열 문자열로 전달.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataCollectionCreateInput) (*mcp.CallToolResult, any, error) {
			payload := map[string]any{
				"name":     input.Name,
				"category": input.Category,
			}
			if input.Description != "" {
				payload["description"] = input.Description
			}
			if input.FileSpecifications != "" {
				var fileSpecs []any
				if err := json.Unmarshal([]byte(input.FileSpecifications), &fileSpecs); err != nil {
					r, _, _ := toolError(fmt.Sprintf("file_specifications가 유효한 JSON 배열이 아닙니다: %v", err))
					return r, nil, nil
				}
				payload["file_specifications"] = fileSpecs
			}

			isDryRun := input.DryRun == nil || *input.DryRun
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			return doCreateWithDryRun(ctx, sc, "/v2/data-collections/", payload, isDryRun)
		})
}
