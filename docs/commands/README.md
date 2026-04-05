# Synapse CLI 커맨드 레퍼런스

Synapse CLI는 Cobra 프레임워크 기반의 Go CLI로, Synapse Backend v2 API를 위한 클라이언트이다.

## 커맨드 트리

```
synapse
├── config
│   ├── add-context      # 새 컨텍스트 추가
│   ├── use-context      # 활성 컨텍스트 전환
│   ├── list-contexts    # 컨텍스트 목록
│   ├── delete-context   # 컨텍스트 삭제
│   ├── set-server       # 서버 URL 설정
│   ├── set-token        # 토큰 설정
│   ├── set-language     # 언어 설정
│   ├── current-context  # 현재 컨텍스트
│   └── view             # 전체 설정 보기
├── login                # 이메일/비밀번호 로그인
├── logout               # 로그아웃
├── tenant
│   ├── list             # 워크스페이스 목록
│   ├── select           # 워크스페이스 선택
│   └── get              # 워크스페이스 상세
├── token
│   ├── list             # 토큰 목록
│   ├── create           # 토큰 생성
│   ├── get              # 토큰 상세
│   └── delete           # 토큰 삭제
├── health               # 서버 상태 확인
├── api                  # 임의 API 호출
└── [16 리소스 커맨드]
    ├── project          # C/R/U/D + permissions, roles, invite
    ├── task             # C/R/U/D + permissions, roles
    ├── assignment       # R + permissions, roles
    ├── review           # R + permissions, roles
    ├── data-collection  # C/R/U/D + permissions, roles, invite
    ├── data-file        # R only
    ├── data-unit        # C/R/U/D + permissions, roles
    ├── experiment       # C/R/U/D + permissions, roles, invite
    ├── gt-dataset       # C/R/U/D + permissions, roles
    ├── gt               # R only
    ├── model            # R only
    ├── job              # R + log 서브커맨드
    ├── plugin           # R + release 서브커맨드
    ├── group            # C/R/U/D + permissions, roles, invite
    ├── workshop         # R + permissions, roles
    └── member           # R + permissions, roles
```

## 글로벌 플래그

모든 커맨드에서 사용할 수 있는 공통 플래그이다.

| 플래그 | 단축 | 설명 | 기본값 |
|--------|------|------|--------|
| `--output` | `-o` | 출력 형식 (`table`, `json`, `yaml`, `ndjson`) | TTY 자동 감지 |
| `--context` | | 활성 컨텍스트 오버라이드 | 설정 파일의 `current_context` |
| `--server` | | 서버 URL 오버라이드 | 컨텍스트 설정값 |
| `--token` | | 인증 토큰 오버라이드 | 컨텍스트 설정값 |
| `--tenant` | | 테넌트(워크스페이스) 코드 오버라이드 | 컨텍스트 설정값 |
| `--verbose` | `-v` | HTTP 요청/응답 헤더 등 상세 출력 | `false` |
| `--dry-run` | | 변경 없이 시뮬레이션 실행 (API 네이티브 `?dry_run=true`) | `false` |
| `--skip-health-check` | | 자동 health check 건너뛰기 | `false` |
| `--no-logo` | | Synapse 로고 숨기기 | `false` |
| `--lang` | | 언어 설정 (`en`, `ko`) | 설정 파일의 `language` |

## 설정 우선순위

CLI 플래그 > 환경 변수 (`SYNAPSE_*`) > 설정 파일 (`config.yaml`)

## 온보딩 플로우

처음 사용 시 다음 순서로 설정한다:

```bash
# 1. 컨텍스트 추가 (서버 URL 등록)
synapse config add-context production --server https://api.synapse.example.com

# 2. 로그인 (이메일/비밀번호)
synapse login

# 3. 워크스페이스 선택
synapse tenant select

# 4. API 사용 가능
synapse project list
```

## 관련 문서

- [config 커맨드](./config.md) -- 설정 관리 (9개 서브커맨드)
- [인증 커맨드](./auth.md) -- login, logout, token 관리
- [리소스 커맨드](./resources.md) -- 16개 리소스의 제네릭 CRUD 패턴
- [api 커맨드](./api.md) -- 임의 API 호출 (escape hatch)
