package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type DataCollectionListInput struct {
	PageAll bool `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type DataCollectionGetInput struct {
	ID string `json:"id" jsonschema:"데이터 컬렉션 ID"`
}

// RegisterDataCollection registers data-collection-related MCP tools.
func RegisterDataCollection(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{Name: "synapse_data_collection_list",
		Description: "Synapse 데이터 컬렉션 목록을 조회한다."},
	func(ctx context.Context, req *mcp.CallToolRequest, input DataCollectionListInput) (*mcp.CallToolResult, any, error) {
		sc, err := newClient(cfg)
		if err != nil {
			r, _, _ := toolError(err.Error())
			return r, nil, nil
		}
		r, _, _ := fetchList(ctx, sc, "/v2/data-collections/", input.PageAll)
		return r, nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "synapse_data_collection_get",
		Description: "Synapse 데이터 컬렉션 상세 정보를 조회한다."},
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
