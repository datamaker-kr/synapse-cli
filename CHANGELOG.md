# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2026-03-19

### Added

- Cobra + Viper 기반 CLI 프레임워크 초기 구성
- OpenAPI 코드젠 파이프라인 (`oapi-codegen`으로 v2 API 197개 client 함수 자동 생성)
- 멀티 환경 프로파일 관리 (`synapse config add-context/use-context/list-contexts/delete-context`)
- 3단계 CLI 진입점 검증 (server → token → tenant, `PersistentPreRunE`)
- 사용자 인증 (`synapse login/logout`, DRF Token + Tenant Access Token)
- 워크스페이스 관리 (`synapse tenant list/get/select`)
- 액세스 토큰 관리 (`synapse token list/create/get/delete`, `--set-config`)
- 서버 헬스 체크 (`synapse health`, `add-context` 시 자동 `/health/` 검증)
- Escape hatch API 커맨드 (`synapse api <METHOD> <PATH>`, stdin 파이프, `--dry-run`)
- 17개 v2 리소스 CRUD 커맨드 (project, task, assignment, review, data-collection, data-file, data-unit, experiment, gt-dataset, gt, model, job, plugin, group, workshop, member)
- 공통 서브리소스 패턴 (permissions, roles, invite)
- 커서 기반 페이지네이션 (`--per-page`, `--cursor`, `--page-all`)
- 출력 포맷 시스템 (table/json/yaml/ndjson, TTY 자동 감지)
- Agent-first 입력 검증 (경로 탐색, 제어 문자, ID 오염, 이중 인코딩 방어)
- Dry-run 모드 (v2 API 네이티브 `?dry_run=true` + 클라이언트 사이드 미리보기)
- 환경 변수 기반 헤드리스 인증 (`SYNAPSE_SERVER/TOKEN/TENANT/ACCESS_TOKEN/CONTEXT`)
- 다국어 지원 (영어/한국어, `go-i18n` + embed FS, `--lang`/`SYNAPSE_LANG`)
- API 요청 시 `Accept-Language` 헤더 자동 포함
- Synapse 로고 ASCII art 브랜딩 (ANSI 색상, `--no-logo`/`SYNAPSE_NO_LOGO`)
- Shell completion (bash/zsh/fish/powershell, Cobra 내장)
- CI/CD 파이프라인 (GitHub Actions ci.yml + release.yml + goreleaser)
- 크로스 플랫폼 빌드 (Linux/macOS/Windows × amd64/arm64)
- v2 API 표준 에러 처리 (`error.code/message/details`, `meta.request_id`)
- OS 네이티브 config 디렉토리 지원 (`os.UserConfigDir()`)
- 코드 포매터 설정 (`goimports` + `gofumpt`)

### Fixed

- login 시 API 에러 응답 nil pointer dereference 수정
- login 응답 코드 2xx 전체를 성공으로 처리 (201 포함)
