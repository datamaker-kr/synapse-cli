package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type DataFileListInput struct {
	DataCollectionID string `json:"data_collection_id" jsonschema:"데이터 컬렉션 ID (필수)"`
	ProcessingStatus string `json:"processing_status,omitempty" jsonschema:"처리 상태 필터 (pending|processing|completed|failed)"`
	Sort             string `json:"sort,omitempty" jsonschema:"정렬 (예: -created, name). 기본: -created"`
	Fields           string `json:"fields,omitempty" jsonschema:"반환 필드 선택 (예: id,name). context window 최적화용"`
	PageAll          bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type DataFilePresignedUploadInput struct {
	DataUnit          int    `json:"data_unit" jsonschema:"data-unit ID (필수)"`
	FileSpecification int    `json:"file_specification" jsonschema:"file-specification ID (필수)"`
	FileName          string `json:"file_name" jsonschema:"업로드할 파일 이름 (필수, 예: car_001.jpg)"`
}

type DataFileConfirmUploadInput struct {
	DataUnit          int `json:"data_unit" jsonschema:"data-unit ID (필수)"`
	FileSpecification int `json:"file_specification" jsonschema:"file-specification ID (필수)"`
}

// RegisterDataFile registers data-file-related MCP tools.
func RegisterDataFile(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_data_file_list",
		Description: "Synapse 데이터 파일 목록을 조회한다. 기본 per_page=50, 최대 200. data_collection_id 필수. 참고: data-unit과 data-file의 연결은 data-unit-file bridge 모델을 통해 관리되지만, v2 API에서 data_unit 기준 필터는 지원하지 않는다. 특정 data-unit에 연결된 파일을 확인하려면 백엔드 관리자에게 문의하거나 data-unit 상세의 meta 정보를 참조할 것.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataFileListInput) (*mcp.CallToolResult, any, error) {
			if input.DataCollectionID == "" {
				r, _, _ := toolError("data_collection_id는 필수입니다.")
				return r, nil, nil
			}
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}

			path := "/v2/data-files/?data_collection=" + url.QueryEscape(input.DataCollectionID)
			if input.ProcessingStatus != "" {
				path += "&processing_status=" + url.QueryEscape(input.ProcessingStatus)
			}

			path = buildListPath(path, input.Sort, input.Fields)
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name: "synapse_data_file_presigned_upload",
		Description: "데이터 파일 업로드용 presigned URL을 발급받는다. " +
			"3단계 워크플로우: " +
			"(1) 이 tool로 presigned URL 발급 → " +
			"(2) MCP 외부에서 PUT <url> -H 'Content-Type: <mime>' --data-binary @<file> 로 실제 업로드 → " +
			"(3) synapse_data_file_confirm_upload 호출하여 서버에 업로드 완료 통지. " +
			"요구 사항: Synapse Backend v2026.1.5+",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataFilePresignedUploadInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			payload := map[string]any{
				"data_unit":          input.DataUnit,
				"file_specification": input.FileSpecification,
				"file_name":          input.FileName,
			}
			return doPostRaw(ctx, sc, "/v2/data-files/presigned-upload/", payload)
		})

	mcp.AddTool(s, &mcp.Tool{
		Name: "synapse_data_file_confirm_upload",
		Description: "presigned URL을 통한 파일 업로드가 완료되었음을 서버에 통지한다. " +
			"presigned-upload → PUT 업로드 → 이 tool 순서로 호출. " +
			"요구 사항: Synapse Backend v2026.1.5+",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input DataFileConfirmUploadInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			payload := map[string]any{
				"data_unit":          input.DataUnit,
				"file_specification": input.FileSpecification,
			}
			return doPostRaw(ctx, sc, "/v2/data-files/confirm-upload/", payload)
		})
}
