# 설정 (Configuration)

synapse-cli는 YAML 기반 멀티 컨텍스트 설정 시스템을 사용합니다. 여러 환경(production, staging 등)을 하나의 설정 파일에서 관리하고, 컨텍스트를 전환하며 사용할 수 있습니다.

## 설정 파일 위치

설정 파일은 다음 우선순위로 디렉토리를 탐색하여 결정됩니다:

| 우선순위 | 경로 | 설명 |
| --- | --- | --- |
| 1 | `$SYNAPSE_CONFIG_DIR` | 환경 변수로 지정한 디렉토리 |
| 2 | `os.UserConfigDir()/synapse` | OS 네이티브 설정 디렉토리 (예: `~/.config/synapse`) |
| 3 | `~/.synapse` | 폴백 디렉토리 |

선택된 디렉토리 내 `config.yaml` 파일이 설정 파일로 사용됩니다. 파일 퍼미션은 보안을 위해 `0600`으로 설정됩니다.

## 설정 파일 구조

```yaml
current_context: prod
language: ko

contexts:
  prod:
    server: https://api.synapse.com
    environment: production
    auth_method: token
    token: drf_xxxx
    tenant_code: my-workspace
    access_token: ""
  staging:
    server: https://staging.synapse.com
    environment: staging
    auth_method: access_token
    token: ""
    tenant_code: ""
    access_token: syn_yyyy
```

### 최상위 필드

| 필드 | 타입 | 설명 |
| --- | --- | --- |
| `current_context` | string | 현재 활성화된 컨텍스트 이름 |
| `language` | string | 출력 언어 (`en` 또는 `ko`) |
| `contexts` | map | 컨텍스트 이름을 키로 하는 설정 맵 |

### ContextConfig 필드

| 필드 | 타입 | 설명 |
| --- | --- | --- |
| `server` | string | Synapse API 서버 URL |
| `environment` | string | 환경 이름 (표시용, 예: `production`, `staging`) |
| `auth_method` | string | 인증 방식 (`token` 또는 `access_token`) |
| `token` | string | DRF Token 인증 시 사용하는 토큰 |
| `tenant_code` | string | DRF Token 인증 시 필요한 워크스페이스 코드 |
| `access_token` | string | Access Token 인증 시 사용하는 토큰 |

## 환경 변수 오버라이드

환경 변수를 사용하여 설정 파일의 값을 오버라이드할 수 있습니다. 환경 변수는 설정 파일보다 높은 우선순위를 가집니다.

| 환경 변수 | 설명 |
| --- | --- |
| `SYNAPSE_CONFIG_DIR` | 설정 디렉토리 경로 |
| `SYNAPSE_CONTEXT` | 활성 컨텍스트 이름 |
| `SYNAPSE_SERVER` | 서버 URL |
| `SYNAPSE_TOKEN` | DRF Token |
| `SYNAPSE_TENANT` | 워크스페이스 코드 |
| `SYNAPSE_ACCESS_TOKEN` | Access Token |
| `SYNAPSE_LANG` | 언어 (`en` 또는 `ko`) |
| `SYNAPSE_NO_LOGO` | 로고 비활성화 (`1`로 설정 시) |

## 설정 우선순위

설정 값은 다음 우선순위로 결정됩니다 (위가 가장 높음):

```
CLI 플래그 (--server, --token 등)
    ↓
환경 변수 (SYNAPSE_SERVER, SYNAPSE_TOKEN 등)
    ↓
설정 파일 (config.yaml)
    ↓
기본값
```

## 컨텍스트 관리

### 컨텍스트 추가

```bash
synapse config add-context prod --server https://api.synapse.com
```

### 컨텍스트 전환

```bash
synapse config use-context staging
```

### 현재 설정 확인

```bash
synapse config view
```

### 일시적 컨텍스트 오버라이드

CLI 플래그를 사용하여 단일 명령에 대해 컨텍스트를 오버라이드할 수 있습니다:

```bash
synapse project list --context staging --server https://other.synapse.com
```
