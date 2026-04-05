---
version: 0.1.0
cli_version: ">=0.0.1"
mcp_server: synapse
---

## Synapse CLI — Claude용 사용 가이드

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

### 주의 사항

- `synapse_login` tool은 보안상 안내 메시지만 반환. 실제 로그인은 터미널에서 직접 수행
- 프로덕션 컨텍스트에서는 더 신중하게 동작
- API 에러 발생 시 에러 메시지에 해결 방법이 포함됨 (401 → 로그인 안내, 404 → ID 확인)
