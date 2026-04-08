# Synapse v2 API 정책

이 문서는 Synapse v2 API의 인증, 속도 제한, 페이지네이션, 입력 검증, 에러 처리, dry-run, 리소스 의존성, 권한 모델에 대한 정책을 정의한다.

---

## 인증 (Authentication)

Synapse v2 API는 세 가지 인증 방식을 지원한다.

### Session 인증

브라우저 기반 세션 쿠키를 사용하는 방식. 웹 UI에서 주로 사용된다.

### Token 인증 (DRF Token)

```
Authorization: Token {token}
SYNAPSE-Tenant: {tenant_code}
```

- `Authorization` 헤더에 DRF Token을 전달한다.
- 반드시 `SYNAPSE-Tenant` 헤더로 테넌트 코드를 함께 지정해야 한다.

### Access Token 인증

```
Authorization: Bearer {access_token}
```

- 또는 `SYNAPSE-ACCESS-TOKEN: syn_{token}` 헤더를 사용한다.
- Access Token에는 테넌트 정보가 포함되어 있으므로 별도의 `SYNAPSE-Tenant` 헤더가 불필요하다.
- Access Token이 설정되어 있으면 DRF Token보다 우선한다.

### 테넌트 자동 할당

- Access Token 방식에서는 토큰에 바인딩된 테넌트가 자동으로 할당된다.
- DRF Token 방식에서는 `SYNAPSE-Tenant` 헤더가 필수이며, 해당 테넌트에 대한 접근 권한이 있어야 한다.

---

## 속도 제한 (Rate Limiting)

API 호출이 속도 제한을 초과하면 HTTP 429 응답이 반환된다.

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please retry after a short delay.",
    "status": 429
  }
}
```

- 429 응답을 받으면 잠시 후 재시도해야 한다.
- `Retry-After` 헤더가 포함될 수 있으며, 해당 시간(초) 후에 재시도를 권장한다.

---

## 페이지네이션 (Pagination)

모든 목록(list) 엔드포인트는 커서 기반 페이지네이션을 사용한다.

### 기본 파라미터

| 파라미터 | 설명 | 기본값 | 최대값 |
|----------|------|--------|--------|
| `per_page` | 한 페이지당 항목 수 | 50 | 200 |
| `cursor` | 다음 페이지 커서 (응답의 `next` 필드) | - | - |

### 정렬 (sort)

`sort` 파라미터로 정렬 기준을 지정한다. 여러 필드를 쉼표로 구분하며, `-` 접두사는 내림차순을 의미한다.

```
GET /v2/projects/?sort=-created,name
```

- `-created`: 생성일 내림차순 (최신 먼저)
- `name`: 이름 오름차순

### 필드 선택 (fields)

`fields` 파라미터로 응답에 포함할 필드를 지정한다. 불필요한 데이터를 줄여 응답 크기를 최적화할 수 있다.

```
GET /v2/projects/?fields=id,name,status
```

### 응답 형식

```json
{
  "results": [...],
  "next": "cursor_token_for_next_page",
  "previous": "cursor_token_for_previous_page",
  "count": 150
}
```

---

## 입력 검증 (Input Validation)

Synapse v2 API는 3계층 방어(3-Layer Defense) 전략으로 입력을 검증한다.

### 1계층: Path Parameter 검증

- 경로 파라미터(예: `{id}`, `{slug}`)의 형식과 유효성을 검증한다.
- 유효하지 않은 ID 형식은 즉시 400 응답으로 거부된다.

### 2계층: Query Parameter 이중 인코딩 방지

- 쿼리 파라미터의 이중 URL 인코딩(double-encoding)을 탐지하고 차단한다.
- Agent가 생성한 URL에서 흔히 발생하는 이중 인코딩 문제를 방어한다.

### 3계층: 문자열 제어 문자 차단

- 문자열 입력에서 제어 문자(control characters)를 탐지하고 제거한다.
- NULL 바이트, 줄바꿈 삽입 등의 공격을 방어한다.

---

## 에러 처리 (Error Handling)

### 에러 응답 형식 (Error Envelope)

모든 에러는 통일된 봉투(envelope) 형식으로 반환된다:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable error description.",
    "status": 400,
    "details": [
      {
        "field": "name",
        "message": "This field is required."
      }
    ]
  }
}
```

### 에러 코드

| HTTP 상태 | 에러 코드 | 설명 |
|-----------|-----------|------|
| 400 | `VALIDATION_ERROR` | 요청 데이터 검증 실패 (필수 필드 누락, 잘못된 형식 등) |
| 401 | `AUTHENTICATION_REQUIRED` | 인증 필요 (토큰 없음 또는 만료) |
| 403 | `PERMISSION_DENIED` | 권한 부족 (해당 리소스에 대한 접근 권한 없음) |
| 404 | `NOT_FOUND` | 리소스를 찾을 수 없음 (ID 확인 필요) |
| 409 | `CONFLICT` | 리소스 충돌 (중복 이름, 동시 수정 등) |
| 422 | `UNPROCESSABLE_ENTITY` | 전제 조건 미충족 (의존 리소스 미존재, 상태 불일치 등) |
| 429 | `RATE_LIMIT_EXCEEDED` | 속도 제한 초과 (잠시 후 재시도) |
| 500 | `INTERNAL_ERROR` | 서버 내부 오류 |

---

## Dry-Run

모든 mutation(생성, 수정, 삭제) 엔드포인트는 `?dry_run=true` 쿼리 파라미터를 지원한다.

```
POST /v2/projects/?dry_run=true
Content-Type: application/json

{"name": "new-project", "description": "..."}
```

- dry-run 요청은 입력 검증과 권한 확인을 수행하지만, 실제 데이터를 변경하지 않는다.
- 응답 형식은 실제 요청과 동일하며, dry-run 결과임을 나타내는 메타데이터가 포함된다.
- **Agent는 모든 mutation 전에 반드시 dry-run을 수행해야 한다.** dry-run 결과를 사용자에게 보여주고, 사용자가 확인한 후에만 실제 mutation을 실행한다.

---

## 리소스 의존성 (Resource Dependencies)

Synapse 리소스는 계층적 의존 관계를 가진다. 리소스를 생성할 때는 반드시 상위 리소스가 먼저 존재해야 한다.

### 데이터 계층

```
Tenant → DataCollection → FileSpecification → DataUnit → DataFile
```

- `Tenant`: 최상위 워크스페이스. 모든 리소스의 루트.
- `DataCollection`: 데이터 컬렉션. Tenant에 소속.
- `FileSpecification`: 파일 사양 정의. DataCollection에 소속.
- `DataUnit`: 데이터 유닛. DataCollection에 소속.
- `DataFile`: 데이터 파일. DataUnit에 소속.

### 프로젝트 계층

```
Project → Task → Assignment
```

- `Project`: 프로젝트. Tenant에 소속.
- `Task`: 태스크. Project에 소속.
- `Assignment`: 할당. Task에 소속.

### 생성 순서

리소스 생성 시 반드시 상위 리소스부터 순서대로 생성해야 한다. 상위 리소스가 존재하지 않으면 422 `UNPROCESSABLE_ENTITY` 에러가 반환된다.

---

## 권한 모델 (Permission Model)

Synapse v2 API는 ViewSet 기반 권한 모델을 사용한다.

### TargetModelViewSet

- 직접 대상이 되는 리소스(Project, DataCollection, Experiment 등)에 대한 ViewSet.
- 사용자의 테넌트 멤버십과 리소스 소유권을 기반으로 접근을 제어한다.

### DerivedModelViewSet

- 상위 리소스로부터 파생된 리소스(Task, Assignment, DataUnit 등)에 대한 ViewSet.
- 상위 리소스의 권한을 상속받는다. 예: Task의 권한은 소속 Project의 권한을 따른다.

### AccessLevel 규칙

| AccessLevel | 읽기 | 쓰기 | 삭제 | 관리 |
|-------------|------|------|------|------|
| `viewer` | O | X | X | X |
| `editor` | O | O | X | X |
| `manager` | O | O | O | X |
| `admin` | O | O | O | O |

- 리소스별로 사용자에게 부여된 AccessLevel에 따라 허용되는 작업이 결정된다.
- 테넌트 관리자(admin)는 테넌트 내 모든 리소스에 대해 전체 권한을 가진다.
