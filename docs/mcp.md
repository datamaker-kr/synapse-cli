# MCP (Model Context Protocol) 연동 가이드

## 개요

`synapse mcp` 서브커맨드는 [MCP(Model Context Protocol)](https://modelcontextprotocol.io/) 서버를 stdio 모드로 실행한다. MCP를 지원하는 AI 클라이언트와 연동하면, 자연어로 Synapse 플랫폼을 조작할 수 있다.

```
사용자: "실행 중인 실험 목록 보여줘"
    ↓
AI 클라이언트 (Claude, ChatGPT 등)
    ↓  tool call: synapse_experiment_list
synapse mcp (stdio MCP 서버)
    ↓  HTTP 호출
Synapse Backend v2 API
    ↓
AI가 결과를 해석하여 자연어로 응답
```

> **사전 조건**: `synapse` CLI가 설치되어 있고, `synapse login`으로 인증이 완료되어야 한다.
> 설치 및 인증은 [getting-started.md](./getting-started.md) 참조.

---

## Claude Code 연동

### 방법 1: 프로젝트 `.mcp.json` (권장)

프로젝트 루트에 `.mcp.json` 파일을 생성한다. Claude Code가 자동으로 인식한다.

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

### 방법 2: 글로벌 설정

`~/.claude/settings.json`에 추가하면 모든 프로젝트에서 사용할 수 있다.

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

### 환경 변수로 컨텍스트 지정

특정 환경(staging, production 등)을 지정할 수 있다:

```json
{
  "mcpServers": {
    "synapse": {
      "command": "synapse",
      "args": ["mcp"],
      "env": {
        "SYNAPSE_CONTEXT": "staging"
      }
    }
  }
}
```

### 사용 예시

Claude Code에서 자연어로 질문하면 된다:

```
"현재 실행 중인 실험이 있어?"
"프로젝트 목록 보여줘"
"experiment exp-123의 잡 로그 확인해줘"
"staging 환경으로 전환해줘"
"새 프로젝트 만들어줘 — 이름은 image-classification"
```

---

## Claude Desktop 연동

`claude_desktop_config.json` 파일에 MCP 서버를 등록한다.

| OS | 설정 파일 경로 |
|----|---------------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |

```json
{
  "mcpServers": {
    "synapse": {
      "command": "/usr/local/bin/synapse",
      "args": ["mcp"]
    }
  }
}
```

> **주의**: Claude Desktop에서는 `synapse` 바이너리의 **절대 경로**를 사용해야 한다.
> `which synapse` 또는 `where synapse`로 경로를 확인하자.

설정 후 Claude Desktop을 재시작하면, 대화창 하단에 MCP 도구 아이콘이 나타난다.

---

## ChatGPT (Codex CLI / ChatGPT Desktop) 연동

ChatGPT Desktop은 MCP 프로토콜을 지원한다. `mcp.json` 설정을 통해 연동할 수 있다.

### ChatGPT Desktop (macOS)

`~/.openai/mcp.json` 파일을 생성한다:

```json
{
  "mcpServers": {
    "synapse": {
      "command": "/usr/local/bin/synapse",
      "args": ["mcp"]
    }
  }
}
```

설정 후 ChatGPT Desktop을 재시작하면, 채팅에서 Synapse tool을 사용할 수 있다.

### Codex CLI

Codex CLI에서 MCP 서버를 사용하려면 프로젝트 루트에 `.mcp.json`을 두면 된다 (Claude Code와 동일한 형식):

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

---

## 기타 MCP 클라이언트 연동

`synapse mcp`는 표준 MCP stdio transport를 사용하므로, MCP를 지원하는 모든 클라이언트에서 사용할 수 있다.

### 공통 설정 패턴

모든 MCP 클라이언트의 설정 형식은 거의 동일하다:

```json
{
  "command": "synapse",
  "args": ["mcp"]
}
```

### MCP Inspector (개발/디버깅)

MCP 서버를 검증하려면 공식 Inspector를 사용한다:

```bash
npx @modelcontextprotocol/inspector -- synapse mcp
```

브라우저에서 열리는 UI로 tool 목록 확인, tool 호출 테스트, resource 읽기 등을 할 수 있다.

### Cursor IDE

Cursor Settings → MCP에서 서버를 추가한다:

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

### Windsurf IDE

`~/.codeium/windsurf/mcp_config.json`에 추가한다:

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

---

## 사용 가능한 Tool (21개)

### 읽기 전용 (13개)

| Tool | 설명 | 주요 파라미터 |
|------|------|--------------|
| `synapse_experiment_list` | 실험 목록 | `status`, `page_all` |
| `synapse_experiment_get` | 실험 상세 | `id` |
| `synapse_job_list` | 잡 목록 | `experiment_id`, `page_all` |
| `synapse_job_log` | 잡 로그 | `job_id` |
| `synapse_project_list` | 프로젝트 목록 | `page_all` |
| `synapse_project_get` | 프로젝트 상세 | `id` |
| `synapse_task_list` | 태스크 목록 | `project_id`, `page_all` |
| `synapse_assignment_list` | 할당 목록 | `task_id`, `page_all` |
| `synapse_data_collection_list` | 데이터 컬렉션 목록 | `page_all` |
| `synapse_data_collection_get` | 데이터 컬렉션 상세 | `id` |
| `synapse_data_unit_list` | 데이터 유닛 목록 | `data_collection_id` (필수), `page_all` |
| `synapse_data_unit_get` | 데이터 유닛 상세 | `id` |
| `synapse_data_file_list` | 데이터 파일 목록 | `data_collection_id` (필수), `page_all` |

### 쓰기 (4개)

| Tool | 설명 | 안전장치 |
|------|------|----------|
| `synapse_project_create` | 프로젝트 생성 | `dry_run` 기본 활성화 (`false`로 실행) |
| `synapse_experiment_create` | 실험 생성 | `dry_run` 기본 활성화 (`false`로 실행) |
| `synapse_project_delete` | 프로젝트 삭제 | `force=true` 필수 |
| `synapse_experiment_delete` | 실험 삭제 | `force=true` 필수 |

### Config / Auth (4개)

| Tool | 설명 |
|------|------|
| `synapse_config_current_context` | 현재 활성 컨텍스트 정보 (서버, 인증 상태) |
| `synapse_config_list_contexts` | 모든 컨텍스트 목록 (토큰은 마스킹) |
| `synapse_config_use_context` | 컨텍스트 전환 (staging ↔ production 등) |
| `synapse_login` | 로그인 방법 안내 (보안상 MCP로 직접 로그인 불가) |

### MCP Resource

| URI | 설명 |
|-----|------|
| `synapse://skills/synapse-cli` | Synapse 플랫폼 개요 및 워크플로우 가이드 |

---

## 안전장치

### 쓰기 작업 — Dry-Run 기본 활성화

`create` 계열 tool은 기본적으로 **dry-run 모드**로 동작한다. 실제로 실행하려면 `dry_run=false`를 명시해야 한다.

```
AI: "프로젝트 'my-project'를 생성하겠습니다."
→ synapse_project_create(name="my-project")           # dry-run 결과만 반환
→ synapse_project_create(name="my-project", dry_run=false)  # 실제 생성
```

### 삭제 작업 — Force 필수

`delete` 계열 tool은 `force=true`가 없으면 거부된다. AI가 사용자에게 확인을 요청한 후에만 실행한다.

### 로그인 — Credential 차단

`synapse_login` tool은 이메일/비밀번호를 받지 않는다. 터미널에서 직접 `synapse login`을 실행하라는 안내만 반환한다.

---

## 트러블슈팅

### "Synapse 컨텍스트가 설정되지 않았습니다"

CLI 초기 설정이 필요하다:

```bash
synapse config add-context <name> --server <url>
synapse login
synapse tenant select <code>
```

### "Authentication required"

로그인 세션이 만료되었다:

```bash
synapse login
```

### "workspace not selected"

워크스페이스를 선택해야 한다:

```bash
synapse tenant list
synapse tenant select <code>
```

### MCP 서버가 연결되지 않음

1. `synapse` 바이너리가 PATH에 있는지 확인: `which synapse`
2. MCP 서버가 정상 동작하는지 확인: `echo '{}' | synapse mcp`
3. 설정 파일의 `command` 경로가 정확한지 확인 (Claude Desktop은 절대 경로 필요)
4. 클라이언트를 재시작

### tool 호출 시 에러

MCP tool의 에러 메시지에 해결 방법이 포함되어 있다:
- `401` → "Authentication required. Run: synapse login"
- `403` → "Permission denied."
- `404` → "Resource not found." (ID 확인)
