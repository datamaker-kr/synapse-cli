# 개발 가이드 (Development Guide)

synapse-cli 프로젝트의 개발 환경 설정, 빌드, 테스트, 코드 생성 등의 워크플로우를 설명합니다.

## 사전 요구사항

| 도구 | 용도 |
| --- | --- |
| Go 1.25+ | 언어 런타임 |
| golangci-lint | 정적 분석 및 린트 |
| goimports | import 정렬 |
| gofumpt | 코드 포매팅 |
| oapi-codegen | OpenAPI 스펙에서 HTTP 클라이언트 코드 생성 |

## Makefile 명령어

### 빌드

```bash
make build
```

`bin/synapse` 바이너리를 생성합니다. Git 태그에서 버전 정보를 추출하여 `ldflags`로 주입합니다.

### 테스트

```bash
make test
```

`go test -race -coverprofile`을 실행합니다. Race condition 감지와 커버리지 리포트가 포함됩니다.

### 포매팅

```bash
make fmt
```

`goimports`로 import를 정렬한 후 `gofumpt`로 코드를 포매팅합니다.

### 린트

```bash
make lint
```

`golangci-lint`를 실행하여 정적 분석을 수행합니다.

### 코드 생성

```bash
make generate
```

`oapi-codegen`을 사용하여 `api/synapse-v2-openapi.yaml` OpenAPI 스펙에서 HTTP 클라이언트 코드를 자동 생성합니다. 생성된 코드는 `internal/client/generated/` 디렉토리에 위치하며, 수동으로 수정해서는 안 됩니다.

### 모듈 관리

```bash
make tidy
```

`go mod tidy`와 `go mod verify`를 실행하여 의존성을 정리하고 검증합니다.

### 취약점 검사

```bash
make vulncheck
```

`govulncheck`을 실행하여 알려진 취약점을 검사합니다.

### 설치

```bash
make install
```

`go install`을 실행하여 `$GOPATH/bin`에 바이너리를 설치합니다.

### 릴리스

```bash
make release
```

`goreleaser`를 사용하여 릴리스 빌드를 생성합니다.

## 프로젝트 구조

```
synapse-cli/
├── cmd/synapse/main.go          # 엔트리포인트
├── internal/
│   ├── cmd/                     # Cobra 커맨드 정의
│   ├── client/                  # HTTP 클라이언트 래퍼 (인증 주입, 에러 처리, 페이지네이션)
│   │   └── generated/           # oapi-codegen 자동 생성 코드 (수정 금지)
│   ├── config/                  # YAML 설정 파일 관리, 컨텍스트 전환
│   ├── output/                  # 출력 포매터 (table/json/yaml/ndjson)
│   ├── i18n/                    # 다국어 지원 (en.yaml, ko.yaml)
│   ├── validation/              # 입력 검증 (Agent hallucination 방어)
│   └── branding/                # 로고 ASCII art
├── api/                         # OpenAPI 스펙 (코드 생성 소스)
├── specs/                       # 설계 명세 문서
├── Makefile                     # 빌드, 테스트, 린트 등 자동화
└── go.mod                       # Go 모듈 정의
```

### 레이어 의존 방향

```
cmd/ → client/, config/, output/, validation/
```

`internal/` 패키지 간 순환 의존은 금지됩니다.

## 주요 의존성

| 패키지 | 용도 |
| --- | --- |
| `cobra` | CLI 프레임워크 |
| `go-pretty` | 테이블 출력 |
| `go-isatty` | TTY 감지 (출력 형식 자동 선택) |
| `go-i18n` | 다국어(i18n) 메시지 |
| `oapi-codegen/runtime` | 자동 생성된 HTTP 클라이언트 런타임 |
| `testify` | 테스트 어서션 |
| `yaml.v3` | YAML 인코딩/디코딩 |
| `x/term` | 터미널 관련 유틸리티 |
| `x/text` | 텍스트 처리 |

## 코딩 컨벤션

- Go 표준 프로젝트 레이아웃 준수 (`cmd/`, `internal/`)
- 에러는 `fmt.Errorf("context: %w", err)` 패턴으로 래핑
- 정상 출력은 stdout, 에러/로고 메시지는 stderr
- Cobra 커맨드 함수명은 `newXxxCmd()` 패턴
- 테스트는 Table-Driven Tests 방식, `t.Parallel()` 사용
- 커버리지 목표: `internal/` 80%+, `cmd/` 60%+

## Exit Code

| 코드 | 의미 |
| --- | --- |
| 0 | 성공 |
| 1 | 일반 에러 |
| 2 | 사용법 에러 |
| 3 | 인증 에러 |
| 4 | 네트워크 에러 |

## TDD 워크플로우

이 프로젝트는 Kent Beck의 TDD(Test-Driven Development) 방법론을 따릅니다:

1. 실패하는 테스트를 먼저 작성 (Red)
2. 테스트를 통과시키는 최소한의 코드 구현 (Green)
3. 테스트가 통과한 후 리팩토링 (Refactor)

구조 변경(structural)과 동작 변경(behavioral)은 반드시 별도의 커밋으로 분리합니다.
