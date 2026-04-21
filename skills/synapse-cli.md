---
version: 0.3.0
cli_version: ">=0.2.0"
backend_version: ">=v2026.1.5"
mcp_server: synapse
---

## Synapse CLI — Claude용 사용 가이드

### 호환 버전

- **synapse-cli**: v0.2.0 이상
- **Synapse Backend**: **v2026.1.5+** (Schema Discovery, presigned-upload, generate-tasks 엔드포인트 요구)

### 플랫폼 개요

Synapse는 데이터 중심 ML 워크플로우 관리 플랫폼이다.

**리소스 계층:**
```
Project → Experiment → Job → Model
    ↓
  Task → Assignment → Review
    ↓
  Data Collection → Data Unit / Data File
```

### 인증 상태 확인

MCP tool 사용 전 반드시 인증 상태를 확인한다:
1. `synapse_config_current_context` — 현재 컨텍스트/서버/인증 상태 확인
2. 인증 안 됨이면 → 사용자에게 "터미널에서 `synapse login`을 실행하세요" 안내
3. 컨텍스트 전환 필요 시 → `synapse_config_list_contexts` → `synapse_config_use_context`

### 작업 전 반드시 확인

- 쓰기 작업(create/delete)은 **기본 dry-run 모드**. `dry_run=false`로 실행 전 반드시 사용자 확인
- delete 작업은 `force=true` 필수. 사용자에게 재확인 후 호출
- 현재 활성 컨텍스트(프로덕션/스테이징)를 사용자에게 먼저 알릴 것

### 자주 쓰는 워크플로우

1. **실행 중인 실험 확인**: `synapse_experiment_list` (status=running)
2. **실험의 잡 로그 확인**: `synapse_experiment_get` → `synapse_job_list` (experiment_id) → `synapse_job_log` (job_id)
3. **프로젝트 탐색**: `synapse_project_list` → `synapse_task_list` (project_id) → `synapse_assignment_list` (task_id)
4. **데이터 관리**: `synapse_data_collection_list` → `synapse_data_unit_list` (data_collection_id) / `synapse_data_file_list` (data_collection_id)

### 페이지네이션

대부분의 list tool은 `page_all` 파라미터를 지원한다:
- `page_all=false` (기본): 첫 페이지만 반환
- `page_all=true`: 모든 페이지를 한번에 수집하여 반환 (대량 데이터 주의)

### Dry-Run 필수 정책

모든 mutation(create/delete) 전에 반드시 dry-run을 수행해야 한다:

1. 먼저 `dry_run=true` (기본값)로 호출하여 검증 결과를 확인한다.
2. dry-run 결과를 사용자에게 보여준다.
3. 사용자가 확인한 후에만 `dry_run=false`로 실제 mutation을 실행한다.

**dry-run 없이 바로 mutation을 실행하지 않는다.**

### sort/fields 파라미터 사용법

목록 조회 시 정렬과 필드 선택을 활용하면 효율적으로 데이터를 조회할 수 있다:

- **정렬**: `sort=-created` (최신 먼저), `sort=name` (이름순), `sort=-created,name` (복합 정렬)
- **필드 선택**: `fields=id,name,status` (필요한 필드만 조회하여 응답 크기 최적화)
- **페이지 크기**: `per_page` 기본값 50, 최대 200

### 리소스 생성 순서

리소스는 반드시 상위 리소스부터 순서대로 생성해야 한다:

```
Tenant → DataCollection → FileSpecification → DataUnit → DataFile
```

상위 리소스가 존재하지 않으면 422 에러가 반환된다.

### validation-script 리소스

validation-script는 데이터 검증 스크립트를 관리하는 리소스이다:
- `synapse_validation_script_list`: 검증 스크립트 목록 조회
- `synapse_validation_script_get`: 검증 스크립트 상세 조회

### 에러 코드 참고

| HTTP 상태 | 에러 코드 | 대응 |
|-----------|-----------|------|
| 400 | `VALIDATION_ERROR` | 요청 데이터 확인 (필수 필드, 형식) |
| 401 | `AUTHENTICATION_REQUIRED` | `synapse login` 실행 |
| 403 | `PERMISSION_DENIED` | 권한 확인 |
| 404 | `NOT_FOUND` | ID 확인 |
| 409 | `CONFLICT` | 리소스 충돌 (중복 이름 등) — 기존 리소스 확인 후 재시도 |
| 422 | `UNPROCESSABLE_ENTITY` | 전제 조건 미충족 — 상위 리소스 존재 여부 확인 |
| 429 | `RATE_LIMIT_EXCEEDED` | 속도 제한 초과 — 잠시 후 재시도 |
| 500 | `INTERNAL_ERROR` | 서버 오류 — 관리자 문의 |

### 리소스 생성 워크플로우 (Schema Discovery 기반)

리소스 생성은 반드시 **Schema Discovery → Dry-Run → Execute** 3단계로 진행한다.

#### data-collection 생성

```
1. synapse_schema_file_specifications(category="image")
   → validation_rules, file_specifications 구조 확인
2. synapse_data_collection_create(
     name="...", category="image",
     file_specifications='[{"name":"image_1","file_type":"image","is_required":true,"is_primary":true,"function_type":"main","index":1}]'
   )  # dry_run 기본 true
3. 사용자 확인 후 dry_run=false로 재호출
```

핵심 규칙:
- naming: `{spec_key}_{index}` 형식 (예: `image_1`)
- `is_primary=true` 1개 필수
- `function_type=main` 1개 필수

#### project 생성

```
1. synapse_schema_annotation_configurations(category="image")
   → annotation_types, configuration_schema 확인
2. synapse_project_create(
     title="...", category="image",
     configuration='{"schema_type":"dm_schema","classification":{"bounding_box":{"id":"<UUID>","representativeCodes":[],"classification_schema":[...]}}}',
     data_collection=123
   )  # dry_run 기본 true
3. 사용자 확인 후 dry_run=false로 재호출
```

핵심 규칙:
- `id` 필드는 모두 UUID v4 (Claude가 직접 생성)
- 빈 configuration은 `'{}'` 전달 (필수 필드)
- category에 `time_series` 추가 지원

#### 파일 업로드 (3단계)

```
1. synapse_data_file_presigned_upload(
     data_unit=1, file_specification=1, file_name="car.jpg"
   )  → presigned URL 반환
2. (MCP 외부) PUT <url> -H 'Content-Type: image/jpeg' --data-binary @car.jpg
3. synapse_data_file_confirm_upload(data_unit=1, file_specification=1)
```

#### 태스크 자동 생성

```
1. (전제) DataUnit의 can_generate_task=true 대기 (비동기 파일 처리 완료)
2. synapse_project_generate_tasks(project_id="123")  # dry_run 기본 true
3. 사용자 확인 후 dry_run=false로 재호출
```

### 리소스 의존 순서

```
Tenant → DataCollection → FileSpecification → DataUnit → DataFile
                                                  ↓
                                        Project → Task → Assignment
```

상위 리소스 없이 하위를 생성하면 422 에러 반환.

### 주의 사항

- `synapse_login` tool은 보안상 안내 메시지만 반환. 실제 로그인은 터미널에서 직접 수행
- 프로덕션 컨텍스트에서는 더 신중하게 동작
- API 에러 발생 시 에러 메시지에 해결 방법이 포함됨 (401 → 로그인 안내, 404 → ID 확인)
- Schema Discovery, presigned-upload, generate-tasks tool은 Synapse Backend v2026.1.5+ 필요
