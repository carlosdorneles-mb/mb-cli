# MB CLI - Makefile
# Usage: make [target]

BINARY_NAME := mb
GO_FILES := $(shell find . -type f -name '*.go' -not -path './vendor/*')
DOCS_DIR := docs-site

# Sandbox dir for run-sandbox (optional: override with make run-sandbox SANDBOX_DIR=/path)
SANDBOX_DIR ?= /tmp/mb-sandbox

.PHONY: all build test clean run run-local run-sandbox install tidy deps \
	build-linux build-darwin build-darwin-arm64 cross \
	release-snapshot \
	docs-install docs-dev docs-build docs-preview \
	lint help

help:
	@echo "MB CLI - targets:"
	@echo ""
	@echo "Executar localmente:"
	@echo "  run            build + ./bin/$(BINARY_NAME) (use ARGS=... para argumentos)"
	@echo "  run-local      go run . (sem build; use ARGS=... para argumentos)"
	@echo "  run-sandbox    go run . com config em $(SANDBOX_DIR) (use ARGS=...)"
	@echo ""
	@echo "Build e testes:"
	@echo "  all            tidy, test, build (default)"
	@echo "  build          compile binary to bin/$(BINARY_NAME)"
	@echo "  test           run tests"
	@echo "  test-coverage  tests + coverage report"
	@echo "  clean          remove bin/, coverage, caches"
	@echo ""
	@echo "Outros:"
	@echo "  install        install to GOPATH/bin"
	@echo "  tidy           go mod tidy"
	@echo "  deps           go mod download"
	@echo "  cross          build linux + darwin (amd64/arm64)"
	@echo "  build-linux    binary for Linux amd64"
	@echo "  build-darwin   binary for macOS amd64"
	@echo "  build-darwin-arm64  binary for macOS arm64"
	@echo "  release-snapshot  run goreleaser release --snapshot (test release pipeline locally)"
	@echo "  lint           run golangci-lint (optional)"
	@echo ""
	@echo "Documentação (Docusaurus em $(DOCS_DIR)):"
	@echo "  docs-install   npm install em $(DOCS_DIR)"
	@echo "  docs-dev       servidor de desenvolvimento (npm run start)"
	@echo "  docs-build     gera $(DOCS_DIR)/dist"
	@echo "  docs-preview   serve $(DOCS_DIR)/dist localmente"

# Default target
all: tidy test build

# Build the CLI binary (default: current OS/arch)
build:
	go build -o bin/$(BINARY_NAME) .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts and caches
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache
	rm -rf $(DOCS_DIR)/dist

# Run the CLI: build then execute binary
run: build
	@./bin/$(BINARY_NAME) $(ARGS)

# Run without building (go run .). Use for quick local testing.
# Example: make run-local ARGS="self sync"
run-local:
	@go run . $(ARGS)

# Run with sandbox config dir (does not touch ~/.config/mb).
# Copies examples/plugins into sandbox and creates SANDBOX_DIR if needed.
# Example: make run-sandbox ARGS="self sync"
run-sandbox:
	@mkdir -p $(SANDBOX_DIR)/mb/plugins
	@if [ -d examples/plugins ]; then cp -r examples/plugins/* $(SANDBOX_DIR)/mb/plugins/; fi
	@chmod +x $(SANDBOX_DIR)/mb/plugins/*/run.sh 2>/dev/null || true
	@XDG_CONFIG_HOME=$(SANDBOX_DIR) go run . $(ARGS)
	@rm -rf $(SANDBOX_DIR)/mb/plugins

# Install binary to $GOPATH/bin or $GOBIN
install: build
	go install .

# Tidy Go module
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Cross-compilation
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 .

# Build all supported platforms (Linux + macOS)
cross: build-linux build-darwin build-darwin-arm64

# Run GoReleaser in snapshot mode (no GitHub release; validates pipeline locally).
# Requires: go install github.com/goreleaser/goreleaser@latest
release-snapshot:
	goreleaser release --snapshot

# Documentation (Docusaurus)
docs-install:
	cd $(DOCS_DIR) && npm install

docs-dev:
	cd $(DOCS_DIR) && npm run start

docs-build:
	cd $(DOCS_DIR) && npm run build

docs-preview:
	cd $(DOCS_DIR) && npx serve dist -p 3000

# Lint (requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...
