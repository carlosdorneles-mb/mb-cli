# MB CLI - Makefile
# Usage: make [target]

BINARY_NAME := mb
GO_FILES := $(shell find . -type f -name '*.go' -not -path './vendor/*')
DOCS_DIR := docs

.PHONY: all build test clean tidy deps \
	install install-plugins-examples uninstall-plugins-examples \
	docs-install docs-dev docs-build docs-preview \
	run run-local \
	check-svu release \
	lint format help

help:
	@echo "MB CLI - targets:"
	@echo ""
	@echo "Executar localmente:"
	@echo "  run            	build + ./bin/$(BINARY_NAME). Uso: make run [comandos...] ou ARGS=\"...\" ou ambos"
	@echo "  run-local      	go run . (sem build). Uso: make run-local [comandos...] ou ARGS=\"...\" ou ambos"
	@echo "                  	Ex.: make run-local tools do ARGS=\"--deploy\"  ou  make run-local ARGS=\"tools do --deploy\""
	@echo "  install-plugins-examples    	registra cada plugin em examples/plugins com 'mb plugins add' (não copia arquivos)"
	@echo "  uninstall-plugins-examples  	remove os plugins de exemplo (infra, tools, etc.) do config do usuário"
	@echo ""
	@echo "Build e testes:"
	@echo "  all            tidy, test, build (default)"
	@echo "  build          compile binary to bin/$(BINARY_NAME)"
	@echo "  test           run tests"
	@echo "  test-coverage  tests + coverage report"
	@echo "  clean          remove bin/, coverage, caches"
	@echo ""
	@echo "Outros:"
	@echo "  install        instala o binário do CLI em $GOPATH/bin"
	@echo "  tidy           limpa dependências não usadas"
	@echo "  deps           instala dependências"
	@echo "  update-deps    atualiza todas as dependências"
	@echo "  lint           executa o golangci-lint"
	@echo "  format         formata o código"
	@echo ""
	@echo "Documentação (Docusaurus em $(DOCS_DIR)):"
	@echo "  docs-install   npm install em $(DOCS_DIR)"
	@echo "  docs-dev       servidor de desenvolvimento (npm run start)"
	@echo "  docs-build     gera $(DOCS_DIR)/dist"
	@echo "  docs-preview   serve $(DOCS_DIR)/dist localmente"
	@echo ""
	@echo "Release (versionamento com svu):"
	@echo "  release       interativo: escolhe major/minor/patch (1-3) e faz push da tag"

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
# Uso: make run [comandos...] ou make run ARGS="..." ou ambos (comandos + ARGS).
run: build
	@./bin/$(BINARY_NAME) $(filter-out run,$(MAKECMDGOALS)) $(ARGS)

# Run without building (go run .). Use for quick local testing.
# Uso: make run-local [comandos...] ou make run-local ARGS="..." ou ambos (comandos + ARGS).
run-local:
	@go run . $(filter-out run-local,$(MAKECMDGOALS)) $(ARGS)

# Registra os plugins de exemplo (apenas diretórios diretos em examples/plugins: infra, tools, etc.).
# Executa na raiz do repo: para cada subdir, mb plugins add <path>. Não copia arquivos.
install-plugins-examples:
	@root=$$(pwd); \
	for subdir in examples/plugins/*/; do \
	  [ -d "$$subdir" ] || continue; \
	  abs=$$(cd "$$root/$$subdir" && pwd); \
	  (cd "$$root" && go run . plugins add "$$abs"); \
	done

# Remove os plugins de exemplo do config (mb plugins remove <package>). Usa os mesmos identificadores que install-examples (infra, tools, etc.).
uninstall-plugins-examples:
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
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	pip install pre-commit
	pre-commit install

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy
	go mod verify

# Documentation (Docusaurus)
docs-install:
	cd $(DOCS_DIR) && npm install

docs-dev:
	cd $(DOCS_DIR) && npm run start

docs-build:
	cd $(DOCS_DIR) && npm install && npm run build

docs-preview:
	cd $(DOCS_DIR) && npx serve dist -p 3000

# Release interativo: mostra opções (current -> next) e usuário escolhe 1, 2 ou 3.
release: check-svu
	@current=$$(svu current 2>/dev/null || echo "v0.0.0"); \
	next_major=$$(svu major); \
	next_minor=$$(svu minor); \
	next_patch=$$(svu patch); \
	echo "Escolha o tipo de release:"; \
	echo "  1. Major ($$current -> $$next_major)"; \
	echo "  2. Minor ($$current -> $$next_minor)"; \
	echo "  3. Patch ($$current -> $$next_patch)"; \
	echo ""; \
	printf "Opção (1-3): "; read opt; \
	case "$$opt" in \
	  1) next="$$next_major";; \
	  2) next="$$next_minor";; \
	  3) next="$$next_patch";; \
	  *) echo "Opção inválida."; exit 1;; \
	esac; \
	git tag "$$next" && git push origin "$$next"

# Release (svu: https://github.com/caarlos0/svu)
check-svu:
	@command -v svu >/dev/null 2>&1 || (echo "svu not installed. Install: https://github.com/caarlos0/svu#installation" && exit 1)

# Lint (requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@golangci-lint run --config .golangci.yml ./...

# Format (requires gofmt: go install golang.org/x/tools/cmd/gofmt@latest)
format fmt:
	goimports -l -w .
	golangci-lint fmt --config .golangci.yml ./...
	@golangci-lint run --config .golangci.yml ./... --fix
	@$(MAKE) lint --quiet

# Catch-all: faz com que "make run plugins sync" repasse plugins sync ao binário (não como alvos)
%:
	@:
