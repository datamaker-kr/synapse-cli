package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type DataUnitListInput struct {
	DataCollectionID string `json:"data_collection_id" jsonschema:"데이터 컬렉션 ID (필수)"`
	Sort             string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields           string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name). context window 최적화용"`
	PageAll          bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type DataUnitGetInput struct {
	ID string `json:"id" jsonschema:"데이터 유닛 ID"`
}

// RegisterDataUnit registers data-unit-related MCP tools.
func RegisterDataUnit(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_data_unit_list",
		Description: "Synapse 데이터 유닛 목록을 조회한다. 기본 per_page=50, 최대 200. data_collection_id 필수. 참고: data-unit에 연결된 파일 구성은 data-collection의 file_specifications로 확인할 수 있다.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataUnitListInput) (*mcp.CallToolResult, any, error) {
			if input.DataCollectionID == "" {
				r, _, _ := toolError("data_collection_id는 필수입니다.")
				return r, nil, nil
			}
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/data-units/?data_collection=" + url.QueryEscape(input.DataCollectionID)
			path = buildListPath(path, input.Sort, input.Fields)
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_data_unit_get",
		Description: "Synapse 데이터 유닛 상세 정보를 조회한다. 참고: 이 유닛에 연결된 data-file 목록은 v2 API에서 직접 조회할 수 없다 (data-unit-file bridge가 v2에 미노출). meta 필드의 original_file_name 등으로 원본 파일 정보를 확인할 수 있다.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataUnitGetInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchOne(ctx, sc, "/v2/data-units/"+input.ID+"/")
			return r, nil, nil
		})
}
