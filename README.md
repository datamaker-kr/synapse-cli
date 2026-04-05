```
       ██
      ████
       ██░░ ████████████████
       ░░░░ ██▓▓▓▓▓▓▓▓▓▓▓▓██
            ██▓▓        ▓▓██    ███████╗██╗   ██╗███╗   ██╗ █████╗ ██████╗ ███████╗███████╗
            ██▓▓        ▓▓██    ██╔════╝╚██╗ ██╔╝████╗  ██║██╔══██╗██╔══██╗██╔════╝██╔════╝
            ██▓▓        ▓▓██    ███████╗ ╚████╔╝ ██╔██╗ ██║███████║██████╔╝███████╗█████╗
            ██▓▓        ▓▓██    ╚════██║  ╚██╔╝  ██║╚██╗██║██╔══██║██╔═══╝ ╚════██║██╔══╝
            ██▓▓▓▓▓▓▓▓▓▓▓▓██    ███████║   ██║   ██║ ╚████║██║  ██║██║     ███████║███████╗
            ████████████████░░  ╚══════╝   ╚═╝   ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚══════╝╚══════╝
                            ██
                           ████                                                   CLI
                            ██
```

# synapse-cli

Synapse Backend v2 API를 위한 CLI 클라이언트. Go로 작성되며, Cobra + Viper 기반의 커맨드라인 인터페이스를 제공한다.

## Overview

Synapse는 데이터 중심 ML 워크플로우 관리 플랫폼이다. `synapse-cli`는 Synapse Backend의 v2 API(21개 리소스 그룹)를 터미널에서 사용할 수 있게 해주는 도구로, 인증, 워크스페이스 관리, 프로젝트/데이터/모델 관리, 임의 API 호출 등의 기능을 제공한다.

## Installation

```bash
# Go install
go install github.com/datamaker-kr/synapse-cli/cmd/synapse@latest

# 또는 릴리즈 바이너리 다운로드 (Linux, macOS, Windows)
# https://github.com/datamaker-kr/synapse-cli/releases
```

## Quick Start

```bash
# 1. 서버 환경 추가 (health check 자동 수행)
synapse config add-context production --server https://api.synapse.example.com

# 2. 로그인
synapse login
# Email: user@example.com
# Password: ********
# Login successful! (context: production)
# Workspaces:
#   ws-prod-001  My Workspace
# Select a workspace:
#   synapse tenant select <code>

# 3. 워크스페이스 선택
synapse tenant select ws-prod-001

# 4. 사용 시작
synapse project list
synapse experiment list --output json
```

## Commands

### Core

| Command                       | Description                  |
| ----------------------------- | ---------------------------- |
| `synapse login`               | 이메일/비밀번호로 로그인     |
| `synapse logout`              | 로컬 자격증명 삭제           |
| `synapse health`              | 서버 헬스 체크               |
| `synapse api <METHOD> <PATH>` | 임의 API 호출 (escape hatch) |

### Workspace & Auth

| Command                             | Description            |
| ----------------------------------- | ---------------------- |
| `synapse tenant list`               | 소속 워크스페이스 목록 |
| `synapse tenant select <code>`      | 워크스페이스 전환      |
| `synapse token create`              | 액세스 토큰 생성       |
| `synapse config add-context <name>` | 환경 프로파일 추가     |
| `synapse config use-context <name>` | 환경 전환              |
| `synapse config view`               | 현재 설정 확인         |

### Resources (v2 API)

| Command                   | Operations                                                        |
| ------------------------- | ----------------------------------------------------------------- |
| `synapse project`         | list, create, get, update, delete + permissions/roles/invite/tags |
| `synapse task`            | list, create, get, update, delete                                 |
| `synapse assignment`      | list, get                                                         |
| `synapse review`          | list, get                                                         |
| `synapse data-collection` | list, create, get, update, delete + groups/invite                 |
| `synapse data-file`       | list, get                                                         |
| `synapse data-unit`       | list, create, get, update, delete                                 |
| `synapse experiment`      | list, create, get, update, delete + invite                        |
| `synapse gt-dataset`      | list, create, get, update, delete + versions                      |
| `synapse gt`              | list, get                                                         |
| `synapse model`           | list, get                                                         |
| `synapse job`             | list, get + log                                                   |
| `synapse plugin`          | list, get + release                                               |
| `synapse group`           | list, create, get, update, delete                                 |
| `synapse workshop`        | list, get                                                         |
| `synapse member`          | list, get                                                         |

## AI 연동 (MCP)

`synapse mcp` 서브커맨드로 [MCP](https://modelcontextprotocol.io/) 서버를 실행하여, AI 어시스턴트에서 자연어로 Synapse를 조작할 수 있다.

### 지원 클라이언트

| 클라이언트 | 설정 파일 |
|-----------|----------|
| **Claude Code** | 프로젝트 루트 `.mcp.json` (자동 인식) |
| **Claude Desktop** | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| **ChatGPT Desktop** | `~/.openai/mcp.json` |
| **Cursor** | Cursor Settings → MCP |
| **Windsurf** | `~/.codeium/windsurf/mcp_config.json` |

### 빠른 설정 (Claude Code)

프로젝트에 이미 `.mcp.json`이 포함되어 있어 별도 설정 없이 바로 사용 가능하다:

```json
{
  "mcpServers": {
    "synapse": {
      "command": "synapse",
      "args": ["mcp"]
    }
  }
}
```

### 사용 예시

```
"현재 실행 중인 실험이 있어?"
"프로젝트 목록 보여줘"
"experiment exp-123의 잡 로그 확인해줘"
"새 프로젝트 만들어줘 — 이름은 image-classification"
```

### 제공 Tool (21개)

- **읽기 13개**: experiment, job, project, task, assignment, data-collection, data-unit, data-file의 list/get
- **쓰기 4개**: project/experiment create (dry-run 기본) / delete (force 필수)
- **설정 4개**: config 조회/전환, login 안내 (보안 guardrail)

> 상세 설정 가이드: [docs/mcp.md](./docs/mcp.md)

## Authentication

두 가지 인증 방식을 지원한다.

### DRF Token (Interactive)

```bash
synapse login
# → DRF Token 발급 후 config에 저장
# → 이후 synapse tenant select <code>로 워크스페이스 선택
```

### Tenant Access Token (Automation)

```bash
# CLI에서 토큰 생성 + config에 자동 저장
synapse token create --description "CI token" --set-config

# 또는 환경 변수
export SYNAPSE_ACCESS_TOKEN=syn_aBcDeFgHiJkLmNoPqRsT...
```

## Configuration

설정 파일 위치: OS 네이티브 config 디렉토리

| OS      | 경로                                                |
| ------- | --------------------------------------------------- |
| Linux   | `~/.config/synapse/config.yaml`                     |
| macOS   | `~/Library/Application Support/synapse/config.yaml` |
| Windows | `%APPDATA%\synapse\config.yaml`                     |

```yaml
current_context: production
language: en

contexts:
  production:
    server: https://api.synapse.example.com
    auth_method: token
    token: "abc123..."
    tenant_code: "ws-prod-001"
  staging:
    server: https://staging-api.synapse.example.com
    auth_method: access_token
    access_token: "syn_xxxx"
```

### 설정 우선순위

1. CLI 플래그 (`--server`, `--token`, `--tenant`, `--context`)
2. 환경 변수 (`SYNAPSE_SERVER`, `SYNAPSE_TOKEN`, `SYNAPSE_TENANT`, ...)
3. 설정 파일

### 환경 변수

| Variable               | Description                |
| ---------------------- | -------------------------- |
| `SYNAPSE_CONTEXT`      | 활성 context 이름          |
| `SYNAPSE_SERVER`       | 서버 URL                   |
| `SYNAPSE_TOKEN`        | DRF Token                  |
| `SYNAPSE_ACCESS_TOKEN` | Tenant Access Token        |
| `SYNAPSE_TENANT`       | 워크스페이스 코드          |
| `SYNAPSE_LANG`         | 언어 (en/ko)               |
| `SYNAPSE_CONFIG_DIR`   | Config 디렉토리 오버라이드 |
| `SYNAPSE_NO_LOGO`      | 로고 숨기기                |

## Output Formats

모든 커맨드는 `--output` (`-o`) 플래그로 출력 포맷을 선택할 수 있다.

```bash
synapse project list              # table (기본, TTY)
synapse project list -o json      # JSON (기본, non-TTY/파이프)
synapse project list -o yaml      # YAML
synapse project list -o ndjson    # NDJSON (스트리밍)
```

TTY가 아닌 경우 (파이프) 자동으로 JSON 출력으로 전환된다.

## Global Flags

```
-o, --output string          출력 포맷 (table|json|yaml|ndjson)
    --context string         컨텍스트 오버라이드
    --server string          서버 URL 오버라이드
    --token string           토큰 오버라이드
    --tenant string          워크스페이스 코드 오버라이드
-v, --verbose                상세 출력 (HTTP 요청/응답)
    --dry-run                드라이런 모드
    --skip-health-check      자동 헬스 체크 건너뛰기
    --no-logo                로고 숨기기
    --lang string            언어 (en|ko)
```

## Internationalization

CLI 메시지는 영어(기본)와 한국어를 지원한다.

```bash
synapse --lang ko tenant list
# 또는
export SYNAPSE_LANG=ko
# 또는
synapse config set-language ko
```

## Shell Completion

```bash
# Bash
synapse completion bash > /etc/bash_completion.d/synapse

# Zsh
synapse completion zsh > "${fpath[1]}/_synapse"

# Fish
synapse completion fish > ~/.config/fish/completions/synapse.fish
```

## Development

```bash
# 빌드
make build

# 테스트
make test

# 포매팅 (goimports → gofumpt)
make fmt

# 린트
make lint

# OpenAPI 코드젠
make generate

# 모듈 정리
make tidy
```

## Project Structure

```
synapse-cli/
├── cmd/synapse/main.go              # 엔트리포인트
├── internal/
│   ├── cmd/                         # Cobra 커맨드 정의
│   ├── client/                      # HTTP 클라이언트 래퍼
│   ├── client/generated/            # oapi-codegen 자동 생성 (수정 금지)
│   ├── config/                      # 설정 파일 관리
│   ├── output/                      # 출력 포매터
│   ├── validation/                  # 입력 검증
│   ├── branding/                    # 로고 ASCII art
│   └── i18n/                        # 다국어 메시지
├── api/synapse-v2-openapi.yaml      # OpenAPI spec
├── Makefile
└── .goreleaser.yaml
```

## License

Copyright (c) 2026 Datamaker. All rights reserved. See [LICENSE](LICENSE).
