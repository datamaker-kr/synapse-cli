# api 커맨드

`synapse api`는 v2 API에서 지원하지 않는 엔드포인트나, CLI에 아직 구현되지 않은 API를 호출할 수 있는 escape hatch 커맨드이다. 임의의 HTTP 메서드와 경로를 지정하여 Synapse 서버에 직접 요청을 보낼 수 있다.

## 사용법

```bash
synapse api <METHOD> <PATH> [--data '{"key": "value"}']
```

| 인자/플래그 | 설명 |
|------------|------|
| `METHOD` | HTTP 메서드 (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) |
| `PATH` | API 경로 (예: `/v2/projects/`) |
| `--data` | 요청 본문 (JSON 문자열) |

## 예시

### GET 요청

```bash
# 프로젝트 목록 조회
synapse api GET /v2/projects/

# 특정 프로젝트 조회
synapse api GET /v2/projects/123/

# 쿼리 파라미터 포함
synapse api GET "/v2/projects/?page_size=10&ordering=-created_at"
```

### POST 요청

```bash
# JSON 데이터로 프로젝트 생성
synapse api POST /v2/projects/ --data '{"title": "New Project"}'
```

### PUT/PATCH 요청

```bash
# 프로젝트 수정
synapse api PATCH /v2/projects/123/ --data '{"title": "Updated Title"}'
```

### DELETE 요청

```bash
# 프로젝트 삭제
synapse api DELETE /v2/projects/123/
```

### 파이프 입력

stdin에서 JSON 데이터를 읽을 수 있다.

```bash
# 파이프로 데이터 전달
echo '{"title": "test"}' | synapse api POST /v2/projects/

# 파일에서 데이터 읽기
cat payload.json | synapse api POST /v2/projects/
```

## dry-run 모드

변경을 발생시키는 메서드(`POST`, `PUT`, `PATCH`, `DELETE`)에 `--dry-run` 플래그를 사용하면 실제 변경 없이 요청을 시뮬레이션한다. API 네이티브 `?dry_run=true` 쿼리 파라미터를 사용한다.

```bash
# 생성 시뮬레이션
synapse api POST /v2/projects/ --data '{"title": "Test"}' --dry-run

# 삭제 시뮬레이션
synapse api DELETE /v2/projects/123/ --dry-run
```

## verbose 모드

`--verbose` (또는 `-v`) 플래그를 사용하면 HTTP 요청/응답 헤더를 포함한 상세 정보를 출력한다. 디버깅에 유용하다.

```bash
synapse api GET /v2/projects/ --verbose
```

출력 예시:

```
> GET /v2/projects/ HTTP/1.1
> Host: api.synapse.example.com
> Authorization: Token xxxxx
> SYNAPSE-Tenant: my-workspace

< HTTP/1.1 200 OK
< Content-Type: application/json

{"results": [...], "next": "..."}
```

## 입력 검증

보안을 위해 다음 입력에 대한 검증이 수행된다:

- **경로 순회(path traversal)**: `../` 등의 경로 순회 패턴이 거부된다.
- **제어 문자**: 경로에 제어 문자가 포함되면 거부된다.
- **이중 인코딩(double encoding)**: `%25` 등의 이중 URL 인코딩 패턴이 거부된다.

이러한 검증은 Agent(AI) 사용 시 hallucination으로 인한 의도치 않은 API 호출을 방어하기 위한 설계이다.

## 출력 형식

`-o` 플래그로 출력 형식을 지정할 수 있다. API 응답이 JSON인 경우:

```bash
# 기본 출력 (JSON)
synapse api GET /v2/projects/

# 테이블 형식
synapse api GET /v2/projects/ -o table

# YAML 형식
synapse api GET /v2/projects/ -o yaml
```
