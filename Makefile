BINARY_NAME := synapse
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
GOFLAGS := -trimpath

.PHONY: build test lint generate tidy vulncheck install clean release

## Build binary
build:
	go build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/synapse/

## Run tests with race detection
test:
	go test -race -coverprofile=coverage.out ./...

## Format code (goimports → gofumpt)
fmt:
	goimports -w -local github.com/datamaker-kr/synapse-cli .
	gofumpt -w .

## Run linter (includes goimports/gofumpt checks)
lint:
	golangci-lint run ./...

## Generate code from OpenAPI spec
generate:
	oapi-codegen --config oapi-codegen.yaml api/synapse-v2-openapi.yaml

## Tidy and verify Go modules
tidy:
	go mod tidy
	go mod verify

## Check for known vulnerabilities
vulncheck:
	govulncheck ./...

## Install binary to GOPATH/bin
install:
	go install $(GOFLAGS) $(LDFLAGS) ./cmd/synapse/

## Clean build artifacts
clean:
	rm -rf bin/ coverage.out

## Release (requires goreleaser)
release:
	goreleaser release --clean

## Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build binary to bin/$(BINARY_NAME)"
	@echo "  test       - Run tests with race detection"
	@echo "  lint       - Run golangci-lint"
	@echo "  generate   - Generate code from OpenAPI spec"
	@echo "  tidy       - go mod tidy && go mod verify"
	@echo "  vulncheck  - Check for known vulnerabilities"
	@echo "  install    - Install binary to GOPATH/bin"
	@echo "  clean      - Clean build artifacts"
	@echo "  release    - Create release with goreleaser"
