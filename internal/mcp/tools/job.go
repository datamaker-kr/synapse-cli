package tools

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type JobListInput struct {
	ExperimentID string `json:"experiment_id,omitempty" jsonschema:"실험 ID로 필터링"`
	PageAll      bool   `json:"page_all,omitempty" jsonschema:"true이면 모든 페이지를 조회한다"`
}

type JobLogInput struct {
	JobID string `json:"job_id" jsonschema:"잡 ID"`
}

// RegisterJob registers job-related MCP tools.
func RegisterJob(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_job_list",
		Description: "Synapse 잡 목록을 조회한다. 실험 ID로 필터링 가능.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input JobListInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			path := "/v2/jobs/"
			if input.ExperimentID != "" {
				path += "?experiment=" + url.QueryEscape(input.ExperimentID)
			}
			r, _, _ := fetchList(ctx, sc, path, input.PageAll)
			return r, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "synapse_job_log",
		Description: "Synapse 잡 로그를 조회한다.",
	},
		func(ctx context.Context, req *mcp.CallToolRequest, input JobLogInput) (*mcp.CallToolResult, any, error) {
			sc, err := newClient(cfg)
			if err != nil {
				r, _, _ := toolError(err.Error())
				return r, nil, nil
			}
			r, _, _ := fetchList(ctx, sc, "/v2/job-logs/?job="+url.QueryEscape(input.JobID), false)
			return r, nil, nil
		})
}
