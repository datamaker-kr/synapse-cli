# Synapse v2 API Policy

> 출처: `https://api.test.synapse.sh/api-docs/v2/policy/` (2026-04-21 fetch 기준)
> 최소 호환 백엔드 버전: **Synapse Backend v2026.1.5+**

이 문서는 Synapse v2 Agent-First API의 핵심 정책을 정리한다. CLI/MCP는 이 정책을 준수한다.

## 1. Quick Start (E2E Workflow)

최소 워크플로우:

```
Schema Discovery → DataCollection 생성 → DataUnit 생성 → DataFile 업로드 (presigned) → Project 생성 → Task 생성
```

### Step 0: Schema Discovery (권장)

```bash
# file specification 스키마
GET /v2/schemas/file-specifications/?category=image

# annotation configuration 스키마
GET /v2/schemas/annotation-configurations/?category=image
```

### Step 1: DataCollection 생성

```json
POST /v2/data-collections/
{
  "name": "Vehicle Detection Dataset",
  "category": "image",
  "file_specifications": [
    {
      "name": "image_1",
      "file_type": "image",
      "is_required": true,
      "is_primary": true,
      "function_type": "main",
      "index": 1
    }
  ]
}
```

- `name`, `category` 필수. `file_specifications` optional
- naming 규칙: `{spec_key}_{index}` (예: `image_1`, `pcd_1`)
- `is_primary=true` 1개 필수
- `function_type=main` 1개 필수

### Step 2: DataUnit 생성

```json
POST /v2/data-units/
{ "data_collection": 1, "name": "sample_001" }
```

### Step 3: DataFile 업로드 (3단계 presigned 워크플로우)

```bash
# 3a. presigned URL 발급
POST /v2/data-files/presigned-upload/
{ "data_unit": 1, "file_specification": 1, "file_name": "car_001.jpg" }

# 3b. presigned URL로 직접 업로드 (MCP 외부)
PUT <presigned_url> -H 'Content-Type: image/jpeg' --data-binary @car_001.jpg

# 3c. 업로드 완료 통지
POST /v2/data-files/confirm-upload/
{ "data_unit": 1, "file_specification": 1 }
```

### Step 4: Project 생성

```json
POST /v2/projects/
{
  "title": "Vehicle Detection Project",
  "category": "image",
  "data_collection": 1,
  "configuration": {
    "schema_type": "dm_schema",
    "classification": {
      "bounding_box": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "representativeCodes": [],
        "classification_schema": [
          {
            "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
            "code": "car",
            "name": "Car",
            "value": "#FF0000",
            "is_default": false,
            "customFields": {},
            "attributes": []
          }
        ]
      }
    }
  }
}
```

- `title`, `category`, `configuration` 필수. `data_collection` nullable optional
- `id` 필드는 모두 **UUID v4** 직접 생성 필요
- 빈 configuration은 `{}` 전달

### Step 5: Task 생성

```bash
POST /v2/projects/{id}/generate-tasks/
```

- 전제: DataUnit의 `can_generate_task=True` (비동기 파일 처리 완료 후 자동 설정)

---

## 2. Response Format

### Success Envelope

```json
{
  "data": { ... },
  "meta": {
    "request_id": "req_abc123def456",
    "pagination": {
      "next_cursor": "cD0yMDI2LTAx",
      "previous_cursor": null,
      "per_page": 50
    }
  }
}
```

### Error Envelope

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "This field is required.",
    "details": [
      { "field": "title", "message": "This field is required." }
    ]
  },
  "meta": { "request_id": "req_abc123def456" }
}
```

### Async Job Envelope (202 Accepted)

```json
{
  "data": {
    "job_id": "job_abc123",
    "status": "queued",
    "status_url": "/v2/jobs/job_abc123"
  },
  "meta": { "request_id": "req_abc123def456" }
}
```

### Error Codes

| HTTP | Error Code | 의미 |
|------|------------|------|
| 400 | `VALIDATION_ERROR` | 입력 데이터 검증 실패 |
| 401 | `AUTHENTICATION_REQUIRED` | 인증 필요 |
| 403 | `PERMISSION_DENIED` | 권한 없음 |
| 404 | `NOT_FOUND` | 리소스 없음 |
| 409 | `CONFLICT` | 리소스 충돌 (중복 등) |
| 422 | `UNPROCESSABLE_ENTITY` | 요청 처리 불가 (전제 조건 미충족) |
| 429 | `RATE_LIMIT_EXCEEDED` | 요청 속도 제한 초과 |
| 500 | `INTERNAL_ERROR` | 서버 내부 오류 |

---

## 3. Authentication & Tenant Context

### 인증 방식

| 방식 | 헤더/메커니즘 | 용도 |
|------|----------|------|
| Session | Django session cookie | Browser / Web UI |
| Token | `Authorization: Token <key>` | CLI / Script |
| Access Token | `Authorization: Bearer <jwt>` | OAuth / Agent |

### Tenant Context

- 인증 과정에서 `request.member.tenant`로 자동 설정
- 별도 `X-Tenant-ID` 헤더 불필요
- 모든 queryset이 `request.member.tenant`로 자동 필터링

---

## 4. Pagination, Sorting & Field Selection

### Cursor 기반 Pagination

| 파라미터 | 기본값 | 최대값 | 설명 |
|----------|--------|--------|------|
| `per_page` | 50 | 200 | 페이지당 결과 수 |
| `cursor` | — | — | base64 cursor 토큰 |

### Sorting

```bash
GET /v2/projects/?sort=-created,name
```

- `-` 접두사: 내림차순
- ViewSet마다 `ALLOWED_SORT_FIELDS` whitelist 존재
- 비허용 필드는 silent ignore (에러 아님)
- 기본: `-created`

### Field Selection

```bash
GET /v2/projects/?fields=id,name,created
```

- List Serializer에 `V2FieldSelectSerializerMixin`이 적용된 경우만 동작
- Detail/Create Serializer에는 적용되지 않음

---

## 5. Input Validation Rules

3계층 방어 (`V2InputValidationMixin` 자동 적용).

### Layer 1: Path Parameters (Resource IDs)

`initial()` 단계에서 검증. 거부 패턴: `[?#%&\\/]` 또는 `..`

```
GET /v2/projects/../../../etc/passwd/  → 400
GET /v2/projects/123?extra=1/          → 400
GET /v2/projects/123%00/               → 400
```

### Layer 2: Query Parameters (Double URL Encoding)

`%25xx` 패턴 거부.

```
GET /v2/projects/?status=%2541         → 400
```

### Layer 3: String Fields (Control Characters)

`perform_create()` / `perform_update()` 단계에서 nested dict/list 재귀 검증.

| 거부 | 허용 |
|------|------|
| `\x00`-`\x08` (NULL, BEL 등) | `\n` (LF) |
| `\x0b` (VT) | `\t` (TAB) |
| `\x0c` (FF) | `\r` (CR) |
| `\x0e`-`\x1f` | |

---

## 6. Safety Rails (Dry-Run 모드)

모든 mutation 엔드포인트(POST, PUT, PATCH, DELETE)에서 `?dry_run=true` 지원.

**Agent는 실제 mutation 전 반드시 dry-run을 수행해야 한다.**

### Dry-Run 응답 (create/update)

```json
{
  "data": {
    "dry_run": true,
    "action": "create",
    "validated_fields": ["title", "category", "configuration"]
  },
  "meta": { "request_id": "req_..." }
}
```

| Action | dry_run=true 동작 |
|--------|--------------------|
| `create` | Serializer validation만 수행, DB insert 없음 |
| `update` | Serializer validation만 수행, DB update 없음 |
| `partial_update` | partial=True validation만 |
| `destroy` | Permission check만, DB delete 없음 |

---

## 7. Resource Workflows

### 생성 순서 (의존성)

```
Tenant → DataCollection → FileSpecification → DataUnit → DataFile
                                                  ↓
                                        Project → Task → Assignment
```

### 의존성 표

| 리소스 | 의존 | 핵심 제약 |
|---|---|---|
| DataCollection | Tenant | tenant 자동 설정 |
| FileSpecification | DataCollection | 컬렉션 생성 시 함께 지정 |
| DataUnit | DataCollection | data_collection 필수 |
| DataFile | DataUnit, FileSpecification | presigned URL 워크플로우 |
| Project | DataCollection | tenant는 DataCollection.tenant로 자동 override |
| Task | Project, DataUnit | (Project, DataUnit) 쌍당 1개 |
| Assignment | Task | task에 의존 |

### 핵심 제약

- **DataUnit.can_generate_task**: 비동기 파일 처리 완료 후 자동 True. Task 생성 전 True여야 함
- **Project.tenant**: 직접 지정 불가, DataCollection.tenant로 자동 override
- **Task uniqueness**: (Project, DataUnit) 쌍당 1개만
- **Presigned upload**: 3단계 (URL 요청 → 업로드 → 완료 통지)

---

## 8. Permission Model

### ViewSet 타입

| ViewSet | Permission Source | 예시 |
|---|---|---|
| `V2TargetModelViewSet` | 자체 MemberRole | Project, DataCollection, Experiment |
| `V2DerivedModelViewSet` | DerivedHop으로 inherit | Task, DataUnit, Workshop |
| `V2ModelViewSet` | 기본 permission stack | ValidationScript, GroundTruthDataset |
| `V2ReadOnlyModelViewSet` | 기본 permission stack | Plugin, Model, DataFile |

### AccessLevel

| 레벨 | list | retrieve | write |
|------|------|----------|-------|
| `PUBLIC` | Yes | Yes | MemberRole 필요 |
| `PARTIALLY_PUBLIC` | Yes | MemberRole 필요 | MemberRole 필요 |
| `PRIVATE` | MemberRole 필요 | MemberRole 필요 | MemberRole 필요 |

### Admin Bypass

Phase 3에서 admin bypass **제거**. 관리자도 각 리소스에 명시적 MemberRole 필요.

---

## 9. Configuration Schema (Project)

### Schema Type

```json
{ "schema_type": "dm_schema" }
```

| 값 | 설명 |
|----|------|
| `dm_schema` | Synapse 기본 (default) |
| `json_schema` | JSON Schema 기반 커스텀 |

### Category → Annotation Types

| Category | 지원 Annotation Types |
|----------|----------------------|
| image | annotationGroup, classification, bounding_box, polygon, polyline, keypoint, relation, segmentation |
| video | annotationGroup, classification, segmentation, bounding_box |
| audio | annotationGroup, classification, segmentation |
| text | classification, relation, named_entity |
| pcd | 3d_bounding_box, 3d_segmentation, relation |
| prompt | classification, prompt, answer |

### Widget Types

| Widget | 설명 | options 필요 |
|--------|------|-------------|
| `select` | 단일 선택 dropdown | Yes |
| `radio` | 라디오 버튼 | Yes |
| `multi_select` | 다중 선택 | Yes |
| `text` | 자유 텍스트 | No |

---

## 10. Schema Discovery APIs

### File Specification Schema

```bash
GET /v2/schemas/file-specifications/                # 전체
GET /v2/schemas/file-specifications/?category=pcd   # 카테고리 필터
```

응답 핵심:
- `categories.{category}.file_specifications` — spec_key별 정의
- `file_types` — 카테고리별 지원 확장자
- `function_types` — `["main", "sub", "meta"]`
- `validation_rules` — naming / primary / main_function / index 규칙
- `payload_schema` — required/optional 필드, 예시

### Annotation Configuration Schema

```bash
GET /v2/schemas/annotation-configurations/?category=image
```

응답 핵심:
- `categories.{category}.annotation_types` — 지원 타입 + smart_tools
- `configuration_schema` — schema_type, classification 구조
- `widget_types` — select/radio/multi_select/text
- `validation_rules` — UUID 생성, category 매칭, attribute_options
- `payload_schema` — required/optional 필드

---

## 11. CLI / MCP 활용

본 정책은 `synapse-cli`의 MCP tool에 다음과 같이 반영:

- **Schema Discovery**: `synapse_schema_file_specifications`, `synapse_schema_annotation_configurations`
- **Creation Workflow**: `synapse_data_collection_create`, `synapse_project_create` (모두 dry_run 기본 활성화)
- **Presigned Upload**: `synapse_data_file_presigned_upload`, `synapse_data_file_confirm_upload`
- **Task Generation**: `synapse_project_generate_tasks`
- **Mutation Safety**: `dry_run` 기본 활성화 (모든 create/delete tool)
- **Sort/Fields**: 모든 list tool 지원
