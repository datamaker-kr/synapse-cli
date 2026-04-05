# 출력 형식 (Output Formats)

synapse-cli는 네 가지 출력 형식을 지원하며, 터미널 환경에 따라 자동으로 적절한 형식을 선택합니다.

## 지원 형식

| 형식 | 설명 | 주요 용도 |
| --- | --- | --- |
| `table` | 정렬된 테이블 형태 | 터미널에서 사람이 읽기 위한 용도 |
| `json` | 들여쓰기된 JSON | 프로그래밍 언어에서 파싱, 디버깅 |
| `yaml` | 표준 YAML | 설정 파일 스타일의 가독성 높은 출력 |
| `ndjson` | 한 줄에 하나의 JSON 객체 | 스트리밍, 파이프라인 처리 |

## 자동 감지

출력 형식을 명시적으로 지정하지 않으면 stdout이 TTY(터미널)인지에 따라 자동으로 결정됩니다:

- **TTY (터미널)**: `table` 형식으로 출력
- **Non-TTY (파이프, 리다이렉션)**: `json` 형식으로 출력

이 자동 감지 덕분에 터미널에서는 읽기 쉬운 테이블을, 스크립트에서는 파싱하기 쉬운 JSON을 별도 설정 없이 받을 수 있습니다.

```bash
# 터미널에서 실행 → table
synapse project list

# 파이프로 전달 → json
synapse project list | jq '.[]'
```

## 형식 지정

`-o` 또는 `--output` 플래그로 출력 형식을 명시적으로 지정할 수 있습니다:

```bash
synapse project list -o json
synapse project list -o yaml
synapse project list -o ndjson
synapse project list -o table
```

## 각 형식 상세

### Table

[go-pretty](https://github.com/jedib0t/go-pretty) 라이브러리를 사용하여 `StyleLight` 스타일의 테이블을 렌더링합니다. 각 리소스의 `ResourceDef.ListCols`에 정의된 컬럼이 표시됩니다.

```
+----------+------------------+--------+
| ID       | TITLE            | STATUS |
+----------+------------------+--------+
| abc123   | My Project       | active |
| def456   | Another Project  | draft  |
+----------+------------------+--------+
```

### JSON

들여쓰기가 적용된 JSON으로 출력됩니다. 전체 응답 구조가 보존됩니다.

```json
[
  {
    "id": "abc123",
    "title": "My Project",
    "status": "active"
  }
]
```

### YAML

표준 YAML 인코딩으로 출력됩니다.

```yaml
- id: abc123
  title: My Project
  status: active
```

### NDJSON (Newline Delimited JSON)

한 줄에 하나의 JSON 객체를 출력합니다. 스트리밍 처리나 파이프라인에서 `jq` 등과 함께 사용하기에 적합합니다.

```
{"id":"abc123","title":"My Project","status":"active"}
{"id":"def456","title":"Another Project","status":"draft"}
```

## 페이지네이션과 --page-all

리스트 명령은 커서 기반 페이지네이션을 지원합니다.

| 플래그 | 설명 |
| --- | --- |
| `--per-page` | 페이지당 항목 수 |
| `--cursor` | 다음 페이지 커서 |
| `--page-all` | 모든 페이지를 순회하여 전체 결과 출력 |

`--page-all`은 모든 페이지를 스트리밍하므로 `ndjson` 형식과 함께 사용하면 대량의 데이터를 효율적으로 처리할 수 있습니다:

```bash
# 모든 프로젝트의 제목만 추출
synapse project list --page-all -o ndjson | jq '.title'
```

## 사용 예시

```bash
# 터미널에서 테이블로 보기
synapse project list

# JSON으로 출력
synapse project list -o json

# 모든 페이지를 NDJSON으로 스트리밍 + jq로 필터링
synapse project list --page-all -o ndjson | jq '.title'

# 특정 리소스를 YAML로 보기
synapse project get abc123 -o yaml

# 파일로 저장
synapse project list -o json > projects.json
```
