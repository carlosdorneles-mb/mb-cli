---
sidebar_position: 3
---

# Gestão de plugins

## Diretório de plugins

O MB usa apenas um diretório, derivado de `os.UserConfigDir()`:

- **Linux**: `~/.config/mb/plugins`
- **macOS**: `~/Library/Application Support/mb/plugins`

## Descoberta

O comando `mb self sync` percorre esse diretório em busca de arquivos `manifest.yaml`. Cada manifesto define um plugin (nome, categoria, tipo, script ou binário).

## Cache SQLite

Os plugins encontrados são gravados em `~/.config/mb/cache.db` (ou equivalente no macOS). Esse banco é consultado na inicialização do CLI para montar a árvore de comandos. Por isso é importante rodar `mb self sync` depois de adicionar ou alterar plugins.

## Comandos no terminal

O MB agrupa plugins por **categoria**. Cada categoria vira um subcomando, e cada plugin vira um subcomando dessa categoria:

- Plugin `name: deploy`, `category: infra` → `mb infra deploy`
- Plugin `name: lint`, `category: dev` → `mb dev lint`

## Execução

Ao rodar `mb infra deploy`, o MB localiza o executável/script no cache, monta o ambiente (merge de env), e executa o processo. Você pode usar scripts shell (`type: sh`) ou binários (`type: bin`).

## Variáveis de ambiente

Antes de executar um plugin, o MB mescla:

1. Variáveis do sistema (`os.Environ()`)
2. Arquivo de defaults: `<UserConfigDir>/mb/env.defaults` (e opcionalmente `--env-file`)
3. Variáveis da linha de comando: `--env KEY=VALUE`

A **maior precedência** é a de `--env`. Assim você pode passar segredos só para o processo do plugin, sem deixá-los no histórico do shell.

Comandos para gerenciar defaults: `mb self env list`, `mb self env set KEY [VALUE]`, `mb self env unset KEY`.
