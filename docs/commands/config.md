# config 커맨드

`synapse config`는 CLI 설정을 관리하는 커맨드 그룹이다. 멀티 컨텍스트를 지원하여 여러 서버 환경(production, staging 등)을 손쉽게 전환할 수 있다.

설정 파일은 OS 네이티브 config 디렉토리에 YAML 형식으로 저장되며, 파일 퍼미션은 `0600`으로 유지된다.

## 서브커맨드

### add-context

새 컨텍스트를 추가한다. 서버 URL을 지정하면 자동으로 `/health/` 엔드포인트에 health check를 수행한다.

```bash
# 기본 사용법
synapse config add-context prod --server https://api.synapse.com

# health check를 건너뛰려면 --force 사용
synapse config add-context prod --server https://api.synapse.com --force
```

health check에 실패하면 컨텍스트가 추가되지 않는다. 서버가 아직 준비되지 않은 경우 `--force` 플래그로 건너뛸 수 있다.

### use-context

활성 컨텍스트를 전환한다. 이후 모든 커맨드는 선택된 컨텍스트의 서버, 토큰, 테넌트 설정을 사용한다.

```bash
synapse config use-context staging
```

### list-contexts

등록된 모든 컨텍스트 목록을 출력한다. 현재 활성 컨텍스트는 `*` 마커로 표시된다.

```bash
synapse config list-contexts

# 별칭(alias)
synapse config list
```

출력 예시:

```
  NAME        SERVER                                  AUTH_METHOD
* production  https://api.synapse.example.com          token
  staging     https://staging-api.synapse.example.com  access_token
```

### delete-context

컨텍스트를 삭제한다. 확인 프롬프트가 표시되며, `--force`로 건너뛸 수 있다.

```bash
# 확인 프롬프트 표시
synapse config delete-context old

# 강제 삭제 (확인 없이)
synapse config delete-context old --force
```

### set-server

현재 활성 컨텍스트의 서버 URL을 변경한다.

```bash
synapse config set-server https://new-api.synapse.com
```

변경 시 `/health/` 엔드포인트로 health check를 수행한다.

### set-token

현재 활성 컨텍스트에 인증 토큰을 설정한다. `auth_method`가 자동으로 `token`으로 변경된다.

```bash
synapse config set-token eyJ0b2tlbi...
```

> **주의**: 토큰은 설정 파일에 저장되며, 파일 퍼미션 `0600`으로 보호된다. 토큰 값은 로그나 stdout에 출력되지 않는다.

### set-language

CLI 메시지 언어를 설정한다. `en` (영어) 또는 `ko` (한국어)를 지원한다.

```bash
synapse config set-language ko
```

설정된 언어는 API 요청 시 `Accept-Language` 헤더로도 전달된다.

### current-context

현재 활성 컨텍스트의 이름을 출력한다.

```bash
synapse config current-context
# 출력: production
```

### view

전체 설정 파일(config.yaml)의 내용을 출력한다. 토큰 값은 마스킹되어 표시된다.

```bash
synapse config view
```

출력 예시:

```yaml
current_context: production
language: ko

contexts:
  production:
    server: https://api.synapse.example.com
    environment: production
    auth_method: token
    token: "****..."
    tenant_code: my-workspace
```

## 설정 파일 구조

```yaml
current_context: production    # 활성 컨텍스트 이름
language: en                   # 전역 언어 설정 ("en" | "ko")

contexts:
  production:
    server: https://api.synapse.example.com
    environment: production
    auth_method: token         # "token" | "access_token"
    token: <drf-token>
    tenant_code: <code>
  staging:
    server: https://staging-api.synapse.example.com
    environment: staging
    auth_method: access_token
    access_token: <syn_xxx>
```

## 설정 우선순위

설정 값은 다음 우선순위로 결정된다:

1. CLI 플래그 (`--server`, `--token`, `--context` 등)
2. 환경 변수 (`SYNAPSE_SERVER`, `SYNAPSE_TOKEN` 등)
3. 설정 파일 (`config.yaml`)
