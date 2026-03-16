# MB CLI - Makefile
# Usage: make [target]

BINARY_NAME := mb
GO_FILES := $(shell find . -type f -name '*.go' -not -path './vendor/*')
DOCS_DIR := docs-site

.PHONY: all build test clean run run-local install install-examples uninstall-examples tidy deps \
	docs-install docs-dev docs-build docs-preview \
	lint help

help:
	@echo "MB CLI - targets:"
	@echo ""
	@echo "Executar localmente:"
	@echo "  run            	build + ./bin/$(BINARY_NAME). Uso: make run [args...] ou make run ARGS=\"...\""
	@echo "  run-local      	go run . (sem build). Uso: make run-local [args...] ou make run-local ARGS=\"...\""
	@echo "                  	Flags como -v/-q: use ARGS (ex.: make run-local ARGS=\"tools hello -v\")"
	@echo "  install-examples    	registra cada plugin em examples/plugins com 'mb plugins add' (não copia arquivos)"
	@echo "  uninstall-examples  	remove os plugins de exemplo (infra, tools, etc.) do config do usuário"
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

# Run the CLI: build then execute binary.
# Uso: make run [args...]  ou  make run ARGS="self sync"
run: build
	@./bin/$(BINARY_NAME) $(or $(ARGS),$(filter-out run,$(MAKECMDGOALS)))

# Run without building (go run .). Use for quick local testing.
# Uso: make run-local [args...]  ou  make run-local ARGS="self sync"
run-local:
	@go run . $(or $(ARGS),$(filter-out run-local,$(MAKECMDGOALS)))

# Registra os plugins de exemplo (apenas diretórios diretos em examples/plugins: infra, tools, etc.).
# Executa na raiz do repo: para cada subdir, mb plugins add <path>. Não copia arquivos.
install-examples:
	@root=$$(pwd); \
	for subdir in examples/plugins/*/; do \
	  [ -d "$$subdir" ] || continue; \
	  abs=$$(cd "$$root/$$subdir" && pwd); \
	  (cd "$$root" && go run . plugins add "$$abs"); \
	done

# Remove os plugins de exemplo do config (mb plugins remove <name>). Usa os mesmos nomes que install-examples (infra, tools, etc.).
uninstall-examples:
	@root=$$(pwd); \
	for subdir in examples/plugins/*/; do \
	  [ -d "$$subdir" ] || continue; \
	  name=$$(basename "$$subdir"); \
	  (cd "$$root" && echo y | go run . plugins remove "$$name"); \
	done

# Install binary to $GOPATH/bin or $GOBIN
install: build
	go install .

# Tidy Go module
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

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

# Catch-all: faz com que "make run self sync" repasse self sync ao binário (não como alvos)
%:
	@:
