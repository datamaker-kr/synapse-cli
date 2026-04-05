# Synapse CLI 아키텍처

이 문서는 synapse-cli의 내부 구조, 패키지 간 의존 관계, 핵심 설계 패턴을 설명합니다.

## 1. 패키지 레이어 다이어그램

```
cmd/synapse/main.go                    엔트리포인트, ldflags로 버전 주입
    |
    v
internal/cmd/                          Cobra 커맨드 정의, PersistentPreRunE 검증
    |
    +---> internal/client/             SynapseClient, RawRequest(), 인증 헤더 주입
    |         |
    |         +---> internal/client/generated/   oapi-codegen 자동 생성 HTTP 클라이언트
    |
    +---> internal/config/             Config, ContextConfig, YAML 파일 영속화
    |
    +---> internal/output/             Formatter 인터페이스 (table/json/yaml/ndjson)
    |
    +---> internal/validation/         입력 안전성 검증 (path traversal, control chars)
    |
    +---> internal/i18n/               go-i18n, en.yaml/ko.yaml embed FS
    |
    +---> internal/branding/           ASCII 로고, 버전 표시
```

레이어 의존 방향은 `cmd/ -> client/, config/, output/, validation/`이며, internal 패키지 간 순환 의존은 금지되어 있습니다.

## 2. 패키지 상세

### cmd/synapse/main.go (엔트리포인트)

- `main.version` 변수가 빌드 시 ldflags로 주입됩니다: `-X main.version=$(VERSION)`
- `cmd.SetVersion(version)` 호출 후 `cmd.Execute()`를 실행합니다.
- 에러 발생 시 `os.Exit(1)`로 종료합니다.

### internal/cmd/ (커맨드 계층)

Cobra 기반 CLI 커맨드를 정의합니다.

| 파일 | 역할 |
|---|---|
| `root.go` | 루트 커맨드, 글로벌 플래그, `PersistentPreRunE` 등록, i18n 초기화 |
| `prerun.go` | 3단계 유효성 검증 (ValidationLevel 0~3) |
| `resources.go` | 16개 리소스의 `ResourceDef` 배열 정의, `registerResourceCommands()` |
| `common_crud.go` | `ResourceDef` 기반 CRUD 커맨드 자동 생성 (list/get/create/update/delete) |
| `common_subresources.go` | 하위 리소스 커맨드 (permissions, roles, invite) |
| `helpers.go` | `buildClient()`, `runWithClient()` 헬퍼 |
| `config.go` | `config` 서브 커맨드 (add-context, use-context, set-server 등) |
| `login.go` | `login`, `logout` 커맨드 |
| `tenant.go` | `tenant list`, `tenant get`, `tenant select` 커맨드 |
| `token.go` | `token` CRUD 커맨드 |
| `health.go` | `health` 커맨드 |
| `api.go` | `api` escape hatch 커맨드 (임의 API 호출) |

### internal/client/ (HTTP 클라이언트)

| 구성 요소 | 역할 |
|---|---|
| `SynapseClient` | 생성된 클라이언트를 래핑하며 인증 헤더 주입, 에러 처리를 담당 |
| `authEditor()` | 모든 요청에 인증 헤더를 자동 주입 (Access Token 또는 DRF Token + Tenant) |
| `RawRequest()` | 임의의 HTTP 요청을 인증 헤더와 함께 수행 |
| `HealthCheck()` | 서버 `/health/` 엔드포인트 연결 확인 (5초 타임아웃) |
| `ParsePaginatedResponse()` | v2 API의 `{data, meta}` 형식 응답 파싱 |
| `StreamNDJSON()` | 페이지네이션된 data 배열을 줄 단위 JSON으로 스트리밍 |
| `generated/` | `oapi-codegen`으로 OpenAPI spec에서 자동 생성된 코드 (수동 수정 금지) |

### internal/config/ (설정 관리)

- 설정 파일 경로: `~/.config/synapse/config.yaml` (OS 네이티브 `os.UserConfigDir()`)
- 파일 퍼미션: `0600` (토큰 보호)
- 멀티 컨텍스트 지원: 여러 서버 환경을 `contexts` 맵으로 관리

설정 파일 구조:

```yaml
current_context: production
language: en

contexts:
  production:
    server: https://api.synapse.example.com
    environment: production
    auth_method: token          # "token" | "access_token"
    token: <drf-token>
    tenant_code: <code>
  staging:
    server: https://staging-api.example.com
    auth_method: access_token
    access_token: <syn_xxx>
```

### internal/output/ (출력 포매터)

`Formatter` 인터페이스를 구현하는 4가지 포매터:

| 포매터 | 형식 | 용도 |
|---|---|---|
| `tableFormatter` | 정렬된 텍스트 테이블 | 터미널(TTY) 기본 출력 |
| `jsonFormatter` | 들여쓰기된 JSON | 파이프/리다이렉트 기본 출력 |
| `yamlFormatter` | YAML | `--output yaml` 지정 시 |
| `ndjsonFormatter` | 줄 단위 JSON | 스트리밍 처리, `--page-all -o ndjson` |

`DetectFormat()`은 TTY 여부를 자동 감지하여 기본 형식을 결정합니다:
- TTY -> `table`
- 파이프/리다이렉트 -> `json`

`Formatter` 인터페이스:

```go
type Formatter interface {
    Format(data interface{}) error
    FormatList(items interface{}, columns []Column) error
    FormatError(statusCode int, body []byte) error
}
```

### internal/i18n/ (다국어 지원)

- `go-i18n` 라이브러리와 `embed.FS`를 사용하여 `en.yaml`, `ko.yaml` 메시지 파일을 바이너리에 내장
- 언어 결정 우선순위: `--lang` 플래그 > `SYNAPSE_LANG` 환경 변수 > 설정 파일 `language` > OS 로캘 > `"en"`
- `Accept-Language` 헤더로 서버 측 응답 언어도 제어

### internal/validation/ (입력 검증)

- 서버 URL 유효성 검사 (`ValidateServerURL`)
- path traversal 공격 방어
- 제어 문자(control characters) 차단
- Agent hallucination 방어를 위한 입력 안전성 검증

### internal/branding/ (브랜딩)

- Synapse ASCII 로고 출력 (stderr로 출력하여 stdout 데이터와 분리)
- 버전 정보 표시
- TTY에서만 자동 표시, `--no-logo` 또는 `SYNAPSE_NO_LOGO=1`로 비활성화

## 3. 핵심 설계 패턴

### 3.1 ResourceDef 기반 동적 커맨드 생성

16개 리소스 커맨드가 단일 `ResourceDef` 구조체 배열에서 자동 생성됩니다:

```go
type ResourceDef struct {
    Name      string           // "project"
    Plural    string           // "projects"
    APIPath   string           // "/v2/projects/"
    IDField   string           // "id"
    ListCols  []output.Column  // 테이블 출력용 컬럼 정의
    HasCreate bool
    HasUpdate bool
    HasDelete bool
}
```

`registerResourceCommands()`가 `resourceDefinitions` 배열을 순회하며 각 리소스에 대해:

1. `newResourceCmd(def)` -> 리소스 상위 커맨드 생성
2. 항상 `list`, `get` 서브 커맨드 추가
3. `HasCreate/HasUpdate/HasDelete` 플래그에 따라 `create`, `update`, `delete` 조건부 추가
4. `subresourceMap`에 정의된 하위 리소스(permissions, roles, invite) 추가
5. 특수 하위 커맨드 추가 (job -> log, plugin -> release)

이 패턴으로 코드 중복 없이 16개 리소스 x 최대 5개 액션 = 최대 80개의 커맨드를 생성합니다.

### 3.2 ValidationLevel (3단계 사전 검증)

`PersistentPreRunE`에서 커맨드 경로에 따라 단계적 검증을 수행합니다:

| Level | 상수 | 검증 내용 | 적용 커맨드 |
|---|---|---|---|
| 0 | `ValidationNone` | 검증 없음 | `config`, `version`, `completion` |
| 1 | `ValidationServer` | 서버 URL 설정 확인 | `health`, `login` |
| 2 | `ValidationAuth` | 서버 + 인증 토큰 확인 | `tenant list/select/get` |
| 3 | `ValidationFull` | 서버 + 토큰 + 테넌트 + 헬스 체크 | 나머지 모든 API 커맨드 |

`getValidationLevel()`은 `commandPath()`로 현재 커맨드의 전체 경로를 구한 뒤, `matchPath()`로 각 수준에 해당하는 커맨드인지 판별합니다. 이를 통해 온보딩 과정에서 불필요한 검증 에러를 방지합니다 (예: 서버만 등록된 상태에서 `login` 실행 가능).

### 3.3 buildClient(): 설정 병합 우선순위

`buildClient(cmd)`는 다음 순서로 설정을 병합합니다:

```
CLI 플래그 (--server, --token, --tenant)
    ↓ 우선
환경 변수 (SYNAPSE_SERVER, SYNAPSE_TOKEN, SYNAPSE_TENANT)
    ↓ 우선
설정 파일 (~/.config/synapse/config.yaml)
```

구체적으로:
1. `config.LoadConfig()`로 설정 파일 로드
2. `cfg.ActiveContext()`로 현재 컨텍스트의 `ContextConfig` 획득
3. 커맨드 플래그가 지정되어 있으면 해당 값으로 덮어쓰기
4. `client.NewSynapseClient(ctxCfg, lang)` 호출로 HTTP 클라이언트 생성

### 3.4 커서 기반 페이지네이션과 streamAllPages()

v2 API는 커서 기반 페이지네이션을 사용합니다. 응답 형식:

```json
{
  "data": [...],
  "meta": {
    "request_id": "...",
    "pagination": {
      "next_cursor": "abc123",
      "previous_cursor": null,
      "per_page": 20
    }
  }
}
```

`streamAllPages()`는 `--page-all` 플래그 사용 시 모든 페이지를 자동 순회합니다:

1. 첫 페이지를 요청합니다.
2. 응답을 `ParsePaginatedResponse()`로 파싱합니다.
3. NDJSON 형식이면 `StreamNDJSON()`으로 줄 단위 출력, 그 외는 포매터로 출력합니다.
4. `HasNextPage()`가 true이면 `cursor=<next_cursor>` 쿼리 파라미터를 추가하여 다음 페이지를 요청합니다.
5. 더 이상 다음 페이지가 없을 때까지 반복합니다.

### 3.5 인증 헤더 주입

`SynapseClient.authEditor()`가 모든 HTTP 요청에 인증 헤더를 자동 주입합니다:

- **Access Token 방식**: `SYNAPSE-ACCESS-TOKEN: syn_{token}` (단독 사용, Tenant 헤더 불필요)
- **DRF Token 방식**: `Authorization: Token {token}` + `SYNAPSE-Tenant: {tenant_code}` (두 헤더 필수)

Access Token이 설정되면 DRF Token보다 우선합니다. 추가로 `Accept-Language` 헤더도 자동 주입됩니다.

## 4. 데이터 플로우: `synapse project list` 실행 과정

```
사용자 입력: synapse project list
    |
    v
[1] cmd.Execute()
    root := newRootCmd()        # 루트 커맨드 생성, 글로벌 플래그 등록
    initI18N(root)              # --lang / env / config에서 언어 결정, i18n.Init() 호출
    root.Execute()              # Cobra 커맨드 트리에서 "project list" 매칭
    |
    v
[2] PersistentPreRunE -> preRunCheck()
    getValidationLevel("project list") -> ValidationFull (Level 3)
    |-- 로고 출력 (TTY + --no-logo 미지정 시, stderr)
    |-- config.LoadConfig() -> 설정 파일 로드
    |-- 서버 URL 설정 확인
    |-- 인증 토큰 확인 (Token 또는 AccessToken)
    |-- 테넌트 코드 확인 (DRF Token 방식인 경우)
    |-- client.HealthCheck() -> GET /health/ (3초 타임아웃, 실패 시 경고만)
    |
    v
[3] RunE -> newResourceListCmd(def).RunE
    buildClient(cmd)
    |-- config.LoadConfig() -> ActiveContext()
    |-- 플래그 오버라이드 적용 (--server, --token, --tenant)
    |-- client.NewSynapseClient(ctxCfg, lang) -> SynapseClient 생성
    |
    v
[4] fetchAndFormat(sc, "/v2/projects/", format, cols, stdout)
    sc.RawRequest(ctx, "GET", "/v2/projects/", nil)
    |-- authEditor(): Authorization + SYNAPSE-Tenant 헤더 주입
    |-- HTTP GET https://api.example.com/v2/projects/
    |
    v
[5] 응답 처리
    client.ParsePaginatedResponse(body)  # {data, meta} 파싱
    json.Unmarshal(pr.Data, &items)      # data 배열을 []map[string]interface{}로 변환
    |
    v
[6] 출력
    output.DetectFormat(outputFlag)      # TTY -> "table", 파이프 -> "json"
    tableFormatter.FormatList(items, cols)
    |
    v
    stdout:
    ID    Title              Category    Created
    1     이미지 분류        vision      2025-01-15
    2     텍스트 분석        nlp         2025-02-20
```

## 5. Exit 코드

| 코드 | 의미 |
|---|---|
| 0 | 성공 |
| 1 | 일반 에러 |
| 2 | 사용법 에러 |
| 3 | 인증 에러 |
| 4 | 네트워크 에러 |

## 6. 출력 원칙

- **정상 데이터**: stdout (파이프/리다이렉트 가능)
- **로고, 상태 메시지, 경고**: stderr (데이터 스트림과 분리)
- **토큰**: 절대 stdout/로그에 출력하지 않음 (`MaskToken()` 사용)
