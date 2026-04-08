package tools

import (
	"context"

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
}
