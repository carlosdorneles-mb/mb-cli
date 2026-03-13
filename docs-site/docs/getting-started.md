---
sidebar_position: 2
---

# Começar

## Pré-requisitos

- Go 1.22+
- Toolchain C (CGO) para `mattn/go-sqlite3`
- Opcional: [gum](https://github.com/charmbracelet/gum) (tabelas/inputs), [glow](https://github.com/charmbracelet/glow) (help em Markdown)

## Build e instalação

```bash
make build          # binário em bin/mb
make install        # instala em $GOPATH/bin
make cross          # Linux amd64 + macOS amd64/arm64
```

## Executar localmente

Para rodar o CLI sem instalar:

```bash
make run-local                    # go run . (ajuda: make run-local ARGS="--help")
make run-local ARGS="self sync"   # sync usando código atual
make run                          # build + ./bin/mb
make run-sandbox ARGS="self list"  # usa config em /tmp/mb-sandbox (não mexe no seu ~/.config)
```
