# Synapse CLI

This document provides a brief overview of the Synapse CLI project.

## Overview

Synapse Backend is a web application built with Python and Django. It serves as a backend system that manages data-driven machine learning workflows, from data collection and annotation to model analysis.

## Overall Engineering Principle

### ROLE AND EXPERTISE

You are a senior software engineer who follows Kent Beck's Test-Driven Development (TDD) and Tidy First principles. Your purpose is to guide development following these methodologies precisely.

### CORE DEVELOPMENT PRINCIPLES

- Always follow the TDD cycle: Red → Green → Refactor
- Write the simplest failing test first
- Implement the minimum code needed to make tests pass
- Refactor only after tests are passing
- Follow Beck's "Tidy First" approach by separating structural changes from behavioral changes
- Maintain high code quality throughout development

### TDD METHODOLOGY GUIDANCE

- Start by writing a failing test that defines a small increment of functionality
- Use meaningful test names that describe behavior (e.g., "shouldSumTwoPositiveNumbers")
- Make test failures clear and informative
- Write just enough code to make the test pass - no more
- Once tests pass, consider if refactoring is needed
- Repeat the cycle for new functionality
- When fixing a defect, first write an API-level failing test then write the smallest possible test that replicates the problem then get both tests to pass.

### TIDY FIRST APPROACH

- Separate all changes into two distinct types:
  1. STRUCTURAL CHANGES: Rearranging code without changing behavior (renaming, extracting methods, moving code)
  2. BEHAVIORAL CHANGES: Adding or modifying actual functionality
- Never mix structural and behavioral changes in the same commit
- Always make structural changes first when both are needed
- Validate structural changes do not alter behavior by running tests before and after

### COMMIT DISCIPLINE

- Only commit when:
  1. ALL tests are passing
  2. ALL compiler/linter warnings have been resolved
  3. The change represents a single logical unit of work
  4. Commit messages clearly state whether the commit contains structural or behavioral changes
- Use small, frequent commits rather than large, infrequent ones

### CODE QUALITY STANDARDS

- Eliminate duplication ruthlessly
- Express intent clearly through naming and structure
- Make dependencies explicit
- Keep methods small and focused on a single responsibility
- Minimize state and side effects
- Use the simplest solution that could possibly work

### REFACTORING GUIDELINES

- Refactor only when tests are passing (in the "Green" phase)
- Use established refactoring patterns with their proper names
- Make one refactoring change at a time
- Run tests after each refactoring step
- Prioritize refactorings that remove duplication or improve clarity

### EXAMPLE WORKFLOW

When approaching a new feature:

1. Write a simple failing test for a small part of the feature
2. Implement the bare minimum to make it pass
3. Run tests to confirm they pass (Green)
4. Make any necessary structural changes (Tidy First), running tests after each change
5. Commit structural changes separately
6. Add another test for the next small increment of functionality
7. Repeat until the feature is complete, committing behavioral changes separately from structural ones

Follow this process precisely, always prioritizing clean, well-tested code over quick implementation.

Always write one test at a time, make it run, then improve structure. Always run all the tests (except long-running tests) each time.

---

# synapse-cli Agent Instructions

## Project Overview

synapse-cli는 Synapse Backend v2 API(21개 리소스 그룹)를 위한 Go CLI 클라이언트이다. Cobra + Viper 기반이며, `oapi-codegen`으로 OpenAPI spec에서 HTTP client를 자동 생성한다. Agent-first 설계 원칙을 적용한다.

## Tech Stack

- **Language**: Go (1.22+)
- **CLI Framework**: Cobra + Viper
- **API Client**: `oapi-codegen` v2 자동 생성 (OpenAPI 3.0.3)
- **Config**: OS 네이티브 config 디렉토리 (YAML, multi-context)
- **Output**: table / json / yaml / ndjson (`--output` flag, TTY 자동 감지)
- **Test**: Go stdlib `testing` + `testify/assert` + `net/http/httptest`
- **i18n**: `go-i18n` + embed FS (en/ko)

## Architecture

```
cmd/synapse/main.go               # 엔트리포인트
internal/cmd/                      # Cobra 커맨드 정의
internal/client/                   # HTTP 클라이언트 래퍼 (인증 주입, 에러 처리, 페이지네이션)
internal/client/generated/         # oapi-codegen 자동 생성 코드 (수정 금지)
internal/config/                   # 설정 파일 관리, 컨텍스트 전환
internal/output/                   # 출력 포매터 (table/json/yaml/ndjson)
internal/validation/               # 입력 검증 (Agent hallucination 방어)
internal/branding/                 # 로고 ASCII art
internal/i18n/                     # 다국어 메시지 (en.yaml, ko.yaml)
api/synapse-v2-openapi.yaml        # OpenAPI spec (코드젠 소스)
```

레이어 의존 방향: `cmd/ → client/, config/, output/, validation/`. internal 패키지 간 순환 의존 금지.

## Coding Conventions

- Go 표준 프로젝트 레이아웃 준수 (`cmd/`, `internal/`)
- `internal/` 하위 패키지는 외부 노출하지 않음
- 에러는 `fmt.Errorf("context: %w", err)` 패턴으로 래핑
- 정상 출력은 stdout, 에러/로고 메시지는 stderr
- Exit code: 0 (성공), 1 (일반 에러), 2 (사용법 에러), 3 (인증 에러), 4 (네트워크 에러)
- Cobra 커맨드 함수명: `newXxxCmd()` 패턴
- 테스트: Table-Driven Tests, `_test.go` 동일 패키지, `t.Parallel()` 독립 테스트
- 커버리지 목표: `internal/` 80%+, `cmd/` 60%+

## Authentication Headers

Synapse API 호출 시 인증 헤더 주입 로직 (`internal/client/client.go:authEditor`):

- **Access Token 방식**: `SYNAPSE-ACCESS-TOKEN: syn_{token}` (단독 사용, tenant 헤더 불필요)
- **DRF Token 방식**: `Authorization: Token {token}` + `SYNAPSE-Tenant: {tenant_code}` (두 헤더 필수)
- Access Token이 설정되어 있으면 DRF Token보다 우선
- **Accept-Language**: 선택된 언어에 따라 `Accept-Language: ko` 등 자동 주입

## Config Structure

```yaml
current_context: production
language: en # "en" | "ko"

contexts:
  production:
    server: https://api.synapse.example.com
    environment: production
    auth_method: token # "token" | "access_token"
    token: <drf-token>
    tenant_code: <code>
  staging:
    server: https://staging-api.synapse.example.com
    environment: staging
    auth_method: access_token
    access_token: <syn_xxx>
```

설정 우선순위: CLI flags > env vars (`SYNAPSE_*`) > config file

## CLI Entry Point Validation (3-Level)

모든 API 커맨드 실행 전 `PersistentPreRunE`에서 단계적 검증:

| Level      | 검증                                   | 적용 커맨드                       |
| ---------- | -------------------------------------- | --------------------------------- |
| 0 (None)   | 없음                                   | `config`, `version`, `completion` |
| 1 (Server) | server 확인                            | `health`, `login`                 |
| 2 (Auth)   | server + token                         | `tenant list/select/get`          |
| 3 (Full)   | server + token + tenant + health check | 나머지 모든 API 커맨드            |

온보딩 플로우: `add-context` → `login` → `tenant select` → API 사용 가능

## v2 API Endpoints (21 Resource Groups)

| Resource         | CLI Command               | CRUD                                                              |
| ---------------- | ------------------------- | ----------------------------------------------------------------- |
| Authentication   | `synapse login/logout`    | login, logout                                                     |
| Tenants          | `synapse tenant`          | list, get, select                                                 |
| Tokens           | `synapse token`           | list, create, get, update, delete                                 |
| Projects         | `synapse project`         | list, create, get, update, delete + permissions/roles/invite/tags |
| Tasks            | `synapse task`            | list, create, get, update, delete                                 |
| Assignments      | `synapse assignment`      | list, get                                                         |
| Reviews          | `synapse review`          | list, get                                                         |
| Data Collections | `synapse data-collection` | list, create, get, update, delete + groups/invite                 |
| Data Files       | `synapse data-file`       | list, get                                                         |
| Data Units       | `synapse data-unit`       | list, create, get, update, delete                                 |
| Experiments      | `synapse experiment`      | list, create, get, update, delete + invite                        |
| GT Datasets      | `synapse gt-dataset`      | list, create, get, update, delete + versions                      |
| Ground Truths    | `synapse gt`              | list, get                                                         |
| Models           | `synapse model`           | list, get                                                         |
| Jobs             | `synapse job`             | list, get + logs                                                  |
| Plugins          | `synapse plugin`          | list, get + releases                                              |
| Groups           | `synapse group`           | list, create, get, update, delete                                 |
| Workshops        | `synapse workshop`        | list, get                                                         |
| Members          | `synapse member`          | list, get                                                         |

## Key Design Decisions

- v2 API 전용 (v1 API 미지원). `synapse api` escape hatch로 임의 API 호출 가능
- `oapi-codegen`으로 OpenAPI spec 기반 HTTP 클라이언트 자동 생성 (197 functions)
- Agent-first 설계: 입력 검증, `--dry-run`, 환경 변수 인증, 구조화 출력, NDJSON 스트리밍
- 리소스 커맨드는 단수형: `synapse project`, `synapse model`
- CRUD 액션 패턴 일관성: `list`, `get`, `create`, `update`, `delete`
- 모든 list에 커서 기반 페이지네이션 (`--per-page`, `--cursor`, `--page-all`)
- 모든 create/update에 `--json` 입력 + `--dry-run` (API 네이티브 `?dry_run=true`)
- Config: OS 네이티브 디렉토리 (`os.UserConfigDir()`), 파일 퍼미션 `0600`
- add-context/set-server 시 `/health/` health check 검증

## Global Flags

```
-o, --output string          Output format (table|json|yaml|ndjson)
    --context string         Override active context
    --server string          Override server URL
    --token string           Override auth token
    --tenant string          Override tenant code
-v, --verbose                Verbose output (HTTP request/response)
    --dry-run                Dry run mode
    --skip-health-check      Skip auto health check
    --no-logo                Hide Synapse logo
    --lang string            Language (en|ko)
```

## Environment Variables

| Variable               | Description                |
| ---------------------- | -------------------------- |
| `SYNAPSE_CONTEXT`      | 활성 context 이름          |
| `SYNAPSE_SERVER`       | 서버 URL                   |
| `SYNAPSE_TOKEN`        | DRF Token                  |
| `SYNAPSE_TENANT`       | 워크스페이스 코드          |
| `SYNAPSE_ACCESS_TOKEN` | Tenant Access Token        |
| `SYNAPSE_LANG`         | 언어 (en/ko)               |
| `SYNAPSE_CONFIG_DIR`   | Config 디렉토리 오버라이드 |
| `SYNAPSE_NO_LOGO`      | 로고 숨기기 (1)            |

## Design References

현재 설계 문서: `specs/initiate-synapse-cli-project-with-specs/`

- `requirements.md` — 요구사항 정의 (FR-1 ~ FR-24)
- `specs.md` — 기술 명세 (TS-1 ~ TS-18)
- `plans.md` — 구현 계획 (Step 1 ~ Step 22)
- `synapse-v2-openapi.yaml` — v2 API OpenAPI spec
- `research-cli-design-for-ai-agents.md` — Agent-first CLI 설계 리서치
- `research-go-test-framework.md` — Go 테스트 프레임워크 리서치

이전 설계 blueprint: `specs/architecture-for-initiate-synapse-cli/references/`

## Do NOT

- `internal/client/generated/` 코드를 수동 수정하지 말 것 (`make generate`로 재생성)
- `internal/` 패키지를 외부에 노출하지 말 것
- 토큰을 로그나 stdout에 출력하지 말 것 (`MaskToken()` 사용)
- 설정 파일 퍼미션을 0600 이외로 설정하지 말 것
- 구조 변경(structural)과 동작 변경(behavioral)을 같은 커밋에 섞지 말 것
- 테스트가 실패하는 상태에서 커밋하지 말 것
