package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/datamaker-kr/synapse-cli/internal/config"
)

type UseContextInput struct {
	Name string `json:"name" jsonschema:"전환할 컨텍스트 이름"`
}

type contextInfo struct {
	Name       string `json:"name"`
	Server     string `json:"server"`
	AuthMethod string `json:"auth_method"`
	Token      string `json:"token"`
	TenantCode string `json:"tenant_code"`
	IsCurrent  bool   `json:"is_current"`
}

// RegisterConfig registers config management and login guardrail MCP tools.
func RegisterConfig(s *mcp.Server, cfg *config.Config) {
	mcp.AddTool(s, &mcp.Tool{Name: "synapse_config_current_context",
		Description: "현재 활성 Synapse 컨텍스트 정보를 반환한다."},
	func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		freshCfg, err := config.LoadConfig()
		if err != nil {
			r, _, _ := toolError(fmt.Sprintf("설정 로드 실패: %v", err))
			return r, nil, nil
		}
		ctxCfg, err := freshCfg.ActiveContext()
		if err != nil {
			r, _, _ := toolError(fmt.Sprintf("활성 컨텍스트 없음: %v", err))
			return r, nil, nil
		}
		info := contextInfo{
			Name:       freshCfg.CurrentContext,
			Server:     ctxCfg.Server,
			AuthMethod: ctxCfg.AuthMethod,
			TenantCode: ctxCfg.TenantCode,
			IsCurrent:  true,
		}
		if ctxCfg.Token != "" {
			info.Token = config.MaskToken(ctxCfg.Token)
		}
		if ctxCfg.AccessToken != "" {
			info.Token = config.MaskToken(ctxCfg.AccessToken)
		}
		r, _, _ := toolText(toJSON(info))
		return r, nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "synapse_config_list_contexts",
		Description: "등록된 모든 Synapse 컨텍스트 목록을 반환한다. 토큰은 마스킹 처리."},
	func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		freshCfg, err := config.LoadConfig()
		if err != nil {
			r, _, _ := toolError(fmt.Sprintf("설정 로드 실패: %v", err))
			return r, nil, nil
		}
		var infos []contextInfo
		for name, c := range freshCfg.Contexts {
			info := contextInfo{
				Name:       name,
				Server:     c.Server,
				AuthMethod: c.AuthMethod,
				TenantCode: c.TenantCode,
				IsCurrent:  name == freshCfg.CurrentContext,
			}
			if c.Token != "" {
				info.Token = config.MaskToken(c.Token)
			}
			if c.AccessToken != "" {
				info.Token = config.MaskToken(c.AccessToken)
			}
			infos = append(infos, info)
		}
		r, _, _ := toolText(toJSON(infos))
		return r, nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "synapse_config_use_context",
		Description: "활성 Synapse 컨텍스트를 전환한다."},
	func(ctx context.Context, req *mcp.CallToolRequest, input UseContextInput) (*mcp.CallToolResult, any, error) {
		freshCfg, err := config.LoadConfig()
		if err != nil {
			r, _, _ := toolError(fmt.Sprintf("설정 로드 실패: %v", err))
			return r, nil, nil
		}
		if err := freshCfg.SetContext(input.Name); err != nil {
			r, _, _ := toolError(err.Error())
			return r, nil, nil
		}
		if err := freshCfg.Save(); err != nil {
			r, _, _ := toolError(fmt.Sprintf("설정 저장 실패: %v", err))
			return r, nil, nil
		}
		r, _, _ := toolText(fmt.Sprintf("컨텍스트를 '%s'로 전환했습니다.", input.Name))
		return r, nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "synapse_login",
		Description: "Synapse 로그인 방법을 안내한다. 보안상 MCP를 통한 직접 로그인은 지원하지 않는다."},
	func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		r, _, _ := toolText("보안상 MCP를 통한 로그인은 지원하지 않습니다.\n\n" +
			"터미널에서 다음 명령어를 직접 실행하세요:\n" +
			"  synapse login\n\n" +
			"이미 로그인했다면 synapse_config_current_context로 현재 상태를 확인할 수 있습니다.")
		return r, nil, nil
	})
}
