# 리소스 커맨드

Synapse CLI는 16개 리소스를 `ResourceDef` 구조체 기반의 제네릭 CRUD 패턴으로 관리한다. 모든 리소스 커맨드는 단수형 이름을 사용한다 (예: `synapse project`, `synapse model`).

## 제네릭 CRUD 패턴

### list -- 목록 조회

리소스 목록을 조회한다. 커서 기반 페이지네이션을 지원한다.

```bash
synapse <resource> list [--per-page N] [--cursor CURSOR] [--page-all]
```

| 플래그 | 설명 | 기본값 |
|--------|------|--------|
| `--per-page` | 페이지당 항목 수 | API 기본값 |
| `--cursor` | 페이지네이션 커서 | 없음 (첫 페이지) |
| `--page-all` | 전체 페이지 자동 순회 | `false` |

```bash
# 기본 목록 조회
synapse project list

# 페이지 크기 지정
synapse project list --per-page 50

# 전체 조회 (모든 페이지 순회)
synapse project list --page-all

# JSON 출력
synapse project list -o json

# NDJSON 스트리밍 (대량 데이터)
synapse project list --page-all -o ndjson
```

### get -- 상세 조회

리소스의 상세 정보를 조회한다.

```bash
synapse <resource> get <id>
```

```bash
synapse project get 123
synapse task get 456 -o json
```

### create -- 생성

리소스를 생성한다. `HasCreate`가 `true`인 리소스에서만 사용 가능하다.

```bash
synapse <resource> create --json '{"field": "value"}' [--dry-run]
```

```bash
# 프로젝트 생성
synapse project create --json '{"title": "New Project", "description": "..."}'

# dry-run으로 시뮬레이션 (API 네이티브 ?dry_run=true)
synapse project create --json '{"title": "Test"}' --dry-run
```

`--json` 플래그에 JSON 문자열을 전달한다. 복잡한 데이터는 파일에서 읽을 수도 있다:

```bash
synapse project create --json "$(cat project.json)"
```

### update -- 수정

리소스를 수정한다. `HasUpdate`가 `true`인 리소스에서만 사용 가능하다.

```bash
synapse <resource> update <id> --json '{"field": "value"}' [--dry-run]
```

```bash
# 프로젝트 제목 수정
synapse project update 123 --json '{"title": "Updated Title"}'

# dry-run으로 시뮬레이션
synapse project update 123 --json '{"title": "Test"}' --dry-run
```

### delete -- 삭제

리소스를 삭제한다. `HasDelete`가 `true`인 리소스에서만 사용 가능하다. 확인 프롬프트가 표시되며, `--force`로 건너뛸 수 있다.

```bash
synapse <resource> delete <id> [--force] [--dry-run]
```

```bash
# 확인 프롬프트 표시
synapse project delete 123

# 강제 삭제 (확인 없이)
synapse project delete 123 --force

# dry-run으로 시뮬레이션
synapse project delete 123 --dry-run
```

## 서브 리소스 커맨드

일부 리소스는 추가 서브커맨드를 제공한다.

### permissions -- 권한 조회

리소스에 대한 현재 사용자의 권한을 조회한다.

```bash
synapse <resource> permissions <id>
```

```bash
synapse project permissions 123
```

### roles -- 역할 조회

리소스에 할당된 역할 목록을 조회한다.

```bash
synapse <resource> roles <id>
```

```bash
synapse project roles 123
```

### invite -- 초대

리소스에 사용자를 초대한다.

```bash
synapse <resource> invite <id>
```

```bash
synapse project invite 123
```

## 특수 서브커맨드

### job log -- 작업 로그

작업(job)의 실행 로그를 조회한다.

```bash
synapse job log list <job_id>
```

### plugin release -- 플러그인 릴리스

플러그인의 릴리스 목록을 조회한다.

```bash
synapse plugin release list <plugin_id>
```

## 리소스 일람표

| 리소스 | CLI 커맨드 | API 경로 | C | R | U | D | 서브커맨드 |
|--------|-----------|----------|---|---|---|---|-----------|
| Project | `synapse project` | `/v2/projects/` | O | O | O | O | permissions, roles, invite |
| Task | `synapse task` | `/v2/tasks/` | O | O | O | O | permissions, roles |
| Assignment | `synapse assignment` | `/v2/assignments/` | | O | | | permissions, roles |
| Review | `synapse review` | `/v2/reviews/` | | O | | | permissions, roles |
| Data Collection | `synapse data-collection` | `/v2/data-collections/` | O | O | O | O | permissions, roles, invite |
| Data File | `synapse data-file` | `/v2/data-files/` | | O | | | |
| Data Unit | `synapse data-unit` | `/v2/data-units/` | O | O | O | O | permissions, roles |
| Experiment | `synapse experiment` | `/v2/experiments/` | O | O | O | O | permissions, roles, invite |
| GT Dataset | `synapse gt-dataset` | `/v2/ground-truth-datasets/` | O | O | O | O | permissions, roles |
| Ground Truth | `synapse gt` | `/v2/ground-truths/` | | O | | | |
| Model | `synapse model` | `/v2/models/` | | O | | | |
| Job | `synapse job` | `/v2/jobs/` | | O | | | log |
| Plugin | `synapse plugin` | `/v2/plugins/` | | O | | | release |
| Group | `synapse group` | `/v2/groups/` | O | O | O | O | permissions, roles, invite |
| Workshop | `synapse workshop` | `/v2/workshops/` | | O | | | permissions, roles |
| Member | `synapse member` | `/v2/members/` | | O | | | permissions, roles |

> **범례**: C=Create, R=Read(list/get), U=Update, D=Delete, O=지원

## 출력 형식

모든 리소스 커맨드는 `-o` (또는 `--output`) 플래그로 출력 형식을 지정할 수 있다.

| 형식 | 설명 | 사용 예시 |
|------|------|----------|
| `table` | 사람이 읽기 쉬운 테이블 형식 | 대화형 터미널 |
| `json` | 들여쓰기된 JSON | 스크립트 연동, jq 파이핑 |
| `yaml` | YAML 형식 | 설정 파일 생성 |
| `ndjson` | 줄 단위 JSON (Newline Delimited JSON) | 대량 데이터 스트리밍 |

TTY 환경에서는 기본값이 `table`이고, 파이프 환경에서는 `json`이 기본값이다.
