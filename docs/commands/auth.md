# 인증 커맨드

Synapse CLI는 두 가지 인증 방식을 지원한다:

- **DRF Token**: `synapse login`으로 획득하는 세션 기반 토큰. `Authorization: Token {token}` + `SYNAPSE-Tenant: {tenant_code}` 헤더로 전송된다.
- **Access Token**: `synapse token create`로 생성하는 장기 토큰. `SYNAPSE-ACCESS-TOKEN: syn_{token}` 헤더 단독으로 전송되며, 테넌트 헤더가 필요 없다.

Access Token이 설정되어 있으면 DRF Token보다 우선 사용된다.

## login

이메일/비밀번호로 로그인한다. 대화형 프롬프트가 표시되며, 성공 시 DRF Token을 설정 파일에 저장한다.

```bash
synapse login
```

실행 흐름:

1. 이메일 입력 프롬프트
2. 비밀번호 입력 프롬프트 (입력 내용 숨김)
3. `POST /users/login/` API 호출
4. 응답에서 DRF Token을 추출하여 현재 컨텍스트에 저장
5. `auth_method`를 `token`으로 설정

> **참고**: `login`은 검증 레벨 1(Server)에 해당하므로 서버 URL만 설정되어 있으면 실행 가능하다. 토큰이나 테넌트 설정은 필요하지 않다.

## logout

현재 컨텍스트에서 인증 정보를 모두 제거한다.

```bash
synapse logout
```

제거되는 설정 값:
- `token`
- `access_token`
- `tenant_code`
- `auth_method`

## token

Access Token을 관리하는 커맨드 그룹이다. Access Token은 CI/CD 파이프라인이나 자동화 스크립트에서 비대화형 인증에 사용된다.

### token list

등록된 토큰 목록을 조회한다.

```bash
synapse token list
```

`GET /v2/tokens/` API를 호출한다. 페이지네이션 플래그를 지원한다.

```bash
# 페이지 크기 지정
synapse token list --per-page 50

# 전체 조회
synapse token list --page-all
```

### token create

새 Access Token을 생성한다. `--description` 플래그로 설명을 지정할 수 있다.

```bash
synapse token create --json '{"description": "CI token"}'
```

생성된 토큰 값은 이 시점에서만 확인할 수 있다. 이후에는 `token get`으로 조회해도 토큰 값이 표시되지 않는다.

`--set-config` 플래그를 사용하면 생성된 토큰을 현재 컨텍스트의 `access_token`에 자동 저장한다.

```bash
synapse token create --json '{"description": "CLI token"}' --set-config
```

### token get

토큰의 상세 정보를 조회한다. 보안상 토큰 값 자체는 표시되지 않는다.

```bash
synapse token get <id>
```

`GET /v2/tokens/<id>/` API를 호출한다.

### token delete

토큰을 삭제한다. 확인 프롬프트가 표시되며, `--force`로 건너뛸 수 있다.

```bash
# 확인 프롬프트 표시
synapse token delete <id>

# 강제 삭제 (확인 없이)
synapse token delete <id> --force
```

## tenant

워크스페이스(테넌트)를 관리하는 커맨드 그룹이다. DRF Token 인증 시 API 호출에 테넌트 코드가 필요하다.

### tenant list

접근 가능한 워크스페이스 목록을 조회한다.

```bash
synapse tenant list
```

### tenant select

활성 워크스페이스를 선택한다. 대화형 프롬프트로 목록에서 선택하거나 직접 지정할 수 있다.

```bash
synapse tenant select
```

선택된 워크스페이스의 `tenant_code`가 현재 컨텍스트에 저장된다.

### tenant get

워크스페이스의 상세 정보를 조회한다.

```bash
synapse tenant get <tenant_code>
```

## 인증 흐름 요약

### 대화형 사용 (개발자)

```bash
synapse config add-context prod --server https://api.synapse.com
synapse login                    # DRF Token 획득
synapse tenant select            # 워크스페이스 선택
synapse project list             # API 사용
```

### 비대화형 사용 (CI/CD)

```bash
# 환경 변수로 인증
export SYNAPSE_SERVER=https://api.synapse.com
export SYNAPSE_ACCESS_TOKEN=syn_xxxxx
synapse project list -o json
```

또는 플래그로 직접 지정:

```bash
synapse project list \
  --server https://api.synapse.com \
  --token syn_xxxxx \
  -o json
```
