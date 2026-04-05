# Synapse CLI 시작 가이드

이 문서는 Synapse CLI를 설치하고 첫 번째 API 호출을 수행하기까지의 과정을 안내합니다.

## 1. 설치

### Go install (권장)

Go 1.22 이상이 설치되어 있다면 다음 명령어로 설치할 수 있습니다:

```bash
go install github.com/datamaker-kr/synapse-cli/cmd/synapse@latest
```

설치 후 `$GOPATH/bin`이 `PATH`에 포함되어 있는지 확인하세요.

### 소스에서 빌드

저장소를 클론한 뒤 `make build`로 빌드합니다:

```bash
git clone https://github.com/datamaker-kr/synapse-cli.git
cd synapse-cli
make build
```

빌드가 완료되면 `bin/synapse` 바이너리가 생성됩니다. 이 파일을 `PATH`가 설정된 디렉토리로 복사합니다:

```bash
cp bin/synapse /usr/local/bin/
```

### 설치 확인

```bash
synapse version
```

## 2. 초기 설정: 서버 컨텍스트 추가

Synapse CLI는 여러 서버 환경(production, staging 등)을 컨텍스트로 관리합니다. 설정 파일은 `~/.config/synapse/config.yaml`에 저장됩니다.

첫 번째 컨텍스트를 추가합니다:

```bash
synapse config add-context production --server https://api.synapse.example.com
```

이 명령은 다음 작업을 수행합니다:

1. 서버 URL의 유효성을 검증합니다.
2. `/health/` 엔드포인트로 서버 연결 상태를 자동 확인합니다.
3. 컨텍스트를 설정 파일에 저장하고 활성 컨텍스트로 전환합니다.

서버에 연결할 수 없는 상태에서도 설정을 저장하려면 `--force` 플래그를 사용합니다:

```bash
synapse config add-context staging --server https://staging-api.example.com --force
```

현재 설정 상태를 확인하려면:

```bash
synapse config view
```

## 3. 로그인

`synapse login` 명령으로 이메일과 비밀번호를 입력하여 인증합니다:

```bash
synapse login
```

```
Email: user@example.com
Password: ********
Login successful! (context: production)

Workspaces:
  acme-corp    ACME Corporation
  research     Research Lab

Select a workspace:
  synapse tenant select <code>
```

로그인에 성공하면 DRF Token이 설정 파일에 자동 저장됩니다. 로그인 후에는 사용 가능한 워크스페이스(Tenant) 목록이 안내로 표시됩니다.

### Access Token 방식 (대안)

이메일/비밀번호 대신 Tenant Access Token을 직접 설정할 수도 있습니다. 이 방식은 CI/CD 파이프라인이나 자동화 스크립트에 적합합니다:

```bash
synapse config set-token syn_xxxxxxxxxxxxx
```

Access Token 방식은 별도의 Tenant 선택이 필요 없습니다 (토큰 자체에 Tenant 정보가 포함됨).

## 4. 워크스페이스(Tenant) 선택

DRF Token 방식으로 로그인한 경우, API를 사용하기 전에 작업할 워크스페이스를 선택해야 합니다.

사용 가능한 워크스페이스 목록을 확인합니다:

```bash
synapse tenant list
```

원하는 워크스페이스를 선택합니다:

```bash
synapse tenant select acme-corp
```

선택한 워크스페이스는 설정 파일에 저장되며, 이후 모든 API 호출에 자동 적용됩니다.

## 5. 첫 번째 API 호출

이제 Synapse API를 사용할 준비가 되었습니다. 프로젝트 목록을 조회합니다:

```bash
synapse project list
```

```
ID    Title              Category    Created
1     이미지 분류        vision      2025-01-15
2     텍스트 분석        nlp         2025-02-20
```

실험 목록을 조회합니다:

```bash
synapse experiment list
```

특정 리소스의 상세 정보를 확인합니다:

```bash
synapse project get 1
```

리소스를 생성할 때는 `--json` 플래그로 요청 본문을 전달합니다:

```bash
synapse project create --json '{"title": "새 프로젝트", "category": "vision"}'
```

실제로 생성하기 전에 `--dry-run` 플래그로 유효성을 먼저 확인할 수 있습니다:

```bash
synapse project create --json '{"title": "새 프로젝트"}' --dry-run
```

### 페이지네이션

목록 조회 시 커서 기반 페이지네이션을 지원합니다:

```bash
# 페이지당 10개 결과
synapse project list --per-page 10

# 모든 페이지를 한 번에 조회
synapse project list --page-all
```

## 6. 출력 형식

Synapse CLI는 출력 환경에 따라 자동으로 형식을 결정합니다:

- **터미널(TTY)**: `table` 형식 (사람이 읽기 편한 테이블)
- **파이프/리다이렉트**: `json` 형식 (프로그램 처리에 적합)

`-o` 플래그로 출력 형식을 직접 지정할 수 있습니다:

```bash
# JSON 형식
synapse project list -o json

# YAML 형식
synapse project list -o yaml

# NDJSON 형식 (줄 단위 JSON, 스트리밍 처리에 적합)
synapse project list --page-all -o ndjson
```

파이프와 함께 사용하는 예시:

```bash
# jq로 프로젝트 이름만 추출
synapse project list | jq '.data[].title'

# 모든 프로젝트를 NDJSON으로 스트리밍
synapse project list --page-all -o ndjson | while read -r line; do
  echo "$line" | jq '.title'
done
```

## 7. 셸 자동 완성

Synapse CLI는 Bash, Zsh, Fish 셸의 명령어 자동 완성을 지원합니다.

### Bash

```bash
# 현재 세션에 적용
source <(synapse completion bash)

# 영구 적용
synapse completion bash > /etc/bash_completion.d/synapse
```

### Zsh

```bash
# 현재 세션에 적용
source <(synapse completion zsh)

# 영구 적용
synapse completion zsh > "${fpath[1]}/_synapse"
```

### Fish

```bash
synapse completion fish | source

# 영구 적용
synapse completion fish > ~/.config/fish/completions/synapse.fish
```

## 8. 유용한 글로벌 플래그

모든 명령에서 사용할 수 있는 글로벌 플래그:

| 플래그 | 설명 |
|---|---|
| `-o, --output` | 출력 형식 (`table`, `json`, `yaml`, `ndjson`) |
| `--context` | 활성 컨텍스트 임시 변경 |
| `--server` | 서버 URL 임시 변경 |
| `--token` | 인증 토큰 임시 변경 |
| `--tenant` | 테넌트 코드 임시 변경 |
| `-v, --verbose` | HTTP 요청/응답 상세 출력 |
| `--dry-run` | 실행 없이 유효성 검증만 수행 |
| `--skip-health-check` | 시작 시 자동 헬스 체크 건너뛰기 |
| `--no-logo` | 시작 시 Synapse 로고 숨기기 |
| `--lang` | 표시 언어 변경 (`en`, `ko`) |

## 9. 환경 변수

CLI 플래그 대신 환경 변수로도 설정을 지정할 수 있습니다. CI/CD 환경에서 유용합니다:

```bash
export SYNAPSE_SERVER=https://api.synapse.example.com
export SYNAPSE_TOKEN=your-drf-token
export SYNAPSE_TENANT=acme-corp
synapse project list
```

설정 우선순위: **CLI 플래그 > 환경 변수 > 설정 파일**

| 환경 변수 | 설명 |
|---|---|
| `SYNAPSE_CONTEXT` | 활성 컨텍스트 이름 |
| `SYNAPSE_SERVER` | 서버 URL |
| `SYNAPSE_TOKEN` | DRF Token |
| `SYNAPSE_TENANT` | 워크스페이스 코드 |
| `SYNAPSE_ACCESS_TOKEN` | Tenant Access Token |
| `SYNAPSE_LANG` | 표시 언어 (`en`, `ko`) |
| `SYNAPSE_CONFIG_DIR` | 설정 디렉토리 경로 변경 |
| `SYNAPSE_NO_LOGO` | `1`로 설정 시 로고 숨기기 |

## 요약: 온보딩 플로우

```
1. synapse config add-context <name> --server <url>   # 서버 등록
2. synapse login                                       # 로그인
3. synapse tenant select <code>                        # 워크스페이스 선택
4. synapse project list                                # API 사용 시작
```
