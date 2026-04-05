# 인증 (Authentication)

synapse-cli는 두 가지 인증 방식을 지원합니다. 설정 파일의 `auth_method` 필드에 따라 API 호출 시 전송되는 헤더가 결정됩니다.

## 인증 방식

### 1. DRF Token 인증

이메일/비밀번호 로그인을 통해 DRF Token을 발급받아 사용하는 방식입니다.

```bash
# 로그인하여 토큰 발급
synapse login
# Email: user@example.com
# Password: ********
# 토큰이 config.yaml에 자동 저장됨
```

API 호출 시 다음 두 개의 헤더가 전송됩니다:

```
Authorization: Token <token>
SYNAPSE-Tenant: <tenant_code>
```

DRF Token 방식은 반드시 `tenant_code`가 설정되어 있어야 합니다.

### 2. Access Token 인증

`synapse token create` 명령으로 생성한 Access Token을 사용하는 방식입니다.

```bash
# Access Token 생성
synapse token create --name "my-token"
```

API 호출 시 단일 헤더만 전송됩니다:

```
SYNAPSE-ACCESS-TOKEN: <token>
```

Access Token 방식은 tenant 헤더가 필요하지 않습니다. Access Token 자체에 워크스페이스 정보가 포함되어 있기 때문입니다.

### 인증 방식 설정

설정 파일에서 `auth_method` 필드를 통해 인증 방식을 지정합니다:

```yaml
contexts:
  prod:
    auth_method: token          # DRF Token 방식
    token: drf_xxxx
    tenant_code: my-workspace
  automation:
    auth_method: access_token   # Access Token 방식
    access_token: syn_yyyy
```

## 검증 수준 (Validation Levels)

모든 명령은 실행 전 `PersistentPreRunE`에서 단계적으로 검증됩니다. 명령의 종류에 따라 요구되는 검증 수준이 다릅니다.

| 수준 | 이름 | 검증 항목 | 적용 명령 |
| --- | --- | --- | --- |
| 0 | None | 없음 | `config`, `version`, `completion` |
| 1 | Server | 서버 URL 확인 | `health`, `login` |
| 2 | Auth | 서버 URL + 토큰 | `tenant list`, `tenant select`, `tenant get` |
| 3 | Full | 서버 URL + 토큰 + 테넌트 + 헬스 체크 | 나머지 모든 API 명령 |

### 온보딩 흐름

처음 CLI를 사용할 때 다음 순서로 설정을 진행합니다:

```bash
# 1. 컨텍스트 추가 (서버 URL 설정)
synapse config add-context prod --server https://api.synapse.com

# 2. 로그인 (토큰 발급)
synapse login

# 3. 워크스페이스 선택
synapse tenant select

# 4. API 사용 가능
synapse project list
```

## 토큰 마스킹

보안을 위해 토큰이 출력될 때는 `config.MaskToken()` 함수에 의해 마스킹됩니다. 마지막 4자리만 표시되고 나머지는 `***`로 대체됩니다.

```
***...XXXX
```

`synapse config view` 등의 명령에서 토큰이 표시될 때 이 마스킹이 적용됩니다. 토큰 전체 값은 로그나 표준 출력에 절대 노출되지 않습니다.

## 환경 변수를 통한 인증

CI/CD 환경이나 자동화 스크립트에서는 환경 변수를 통해 인증 정보를 전달할 수 있습니다:

```bash
# DRF Token 방식
export SYNAPSE_SERVER=https://api.synapse.com
export SYNAPSE_TOKEN=drf_xxxx
export SYNAPSE_TENANT=my-workspace

# Access Token 방식
export SYNAPSE_SERVER=https://api.synapse.com
export SYNAPSE_ACCESS_TOKEN=syn_yyyy

# 명령 실행
synapse project list
```
