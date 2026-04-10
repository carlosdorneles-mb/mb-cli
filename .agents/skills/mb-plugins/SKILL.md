---
name: mb-plugins
description: >-
  Covers MB CLI plugin lifecycle (`mb plugins`): sync, add, list, remove, update,
  SQLite cache, manifest scanning, dynamic `mb <categoria>` commands, plugin
  execution, and embedded shell helpers (`internal/infra/shellhelpers`, MB_HELPERS_PATH).
  Use when changing or explaining `mb plugins`, `manifest.yaml`, scanner,
  `plugin_sources`, `plugincmd`, plugin security paths, `groups.yaml`, shell helpers
  embed, or docs under plugins / creating-plugins / plugin-commands / helpers-shell.
---

# MB CLI — `mb plugins` e execução de plugins

## Quando aplicar

- Implementar ou corrigir **`mb plugins`** (add, list, remove, update, sync) ou o **cache SQLite** de comandos.
- Alterar **descoberta** (`Scanner`), **manifest** (`manifest.yaml`), **validação de paths**, **Git** (add remoto / update).
- Trabalhar com **`cli/plugincmd`** (anexar comandos dinâmicos ao root, executar entrypoints, `env_files`, flags).
- Atualizar **documentação** de plugins ou segurança (paths confinados ao pacote).
- Alterar **helpers de shell** embebidos (`internal/infra/shellhelpers`) ou documentação em `helpers-shell.md`.

## Comandos (superfície)

| Comando | Notas |
|--------|--------|
| `mb plugins sync` | Rescan + reconstrói cache; `--no-remove` mantém órfãos no SQLite; com **`mb -v`** o scanner pode emitir debug |
| `mb plugins add <url \| path \| .>` | Remoto: clone + registo; local: `local_path` em `plugin_sources`; `--package`, `--tag`, `--no-remove` |
| `mb plugins list` | Plugins instalados; `--check-updates` (tags Git quando aplicável) |
| `mb plugins remove <package>` | Remove pacote + sync |
| `mb plugins update [package \| --all]` | Remotos; `--all` atualiza todos com GitURL; depois sync |

Aliases do grupo: **`plugin`**, **`p`**, **`extensions`**, **`e`**.

## Onde está o código

| Área | Caminho |
|------|---------|
| Casos de uso | `internal/app/plugins/` (`sync.go`, `add_remote.go`, `add_local.go`, `remove.go`, `update.go`, `runtime.go`, `fsutil.go`) |
| Cobra | `internal/cli/plugins/` (`plugins.go`, `sync.go`, `sync_run.go`, `add.go`, `list.go`, `remove.go`, `update.go`) |
| Descoberta / manifest / Git | `internal/infra/plugins/` (`scanner.go`, `manifest.go`, `manifest_env.go`, `git.go`, `git_service.go`, `layout_validator.go`, `source.go`, `groups.go`, `plugin_leaf_hash.go`) |
| Comandos dinâmicos + execução | `internal/cli/plugincmd/` (`attach.go`, `run.go`, `leaf.go`, …) |
| Modelo / DTOs cache | `internal/domain/plugin/` |
| Persistência cache | `internal/infra/sqlite/` (implementa stores usados no sync) |
| FX | `internal/module/plugins/`, `internal/module/cache/`, `internal/module/executor/` |
| `mb update` (fase plugins) | `internal/cli/update/` chama `cli/plugins` |

Detalhe por ficheiro: [reference.md](reference.md).

## Regras de produto (resumo)

1. **Cache**: metadados em **SQLite** (`ConfigDir/cache.db`); árvore escaneada em **`PluginsDir`** e em cada **`plugin_sources.local_path`**.
2. **`RunSync`**: `scanner.Scan()` → merge com fontes locais → substitui linhas de plugins/categorias/help groups; **`SyncOptions.NoRemove`** evita apagar comandos que desapareceram do disco (órfãos).
3. **Manifest**: entrypoints, `env_files`, `flags` e readme têm de ficar **dentro** do diretório do plugin; manifest inválido → aviso e ignorar essa folha (ver doc técnica).
4. **Após add/remove/update**: fluxos chamam **sync** (com `EmitSuccess` conforme o caso) para refrescar o cache e **shell helpers** (`~/.config/mb/lib/shell`).
5. **Execução**: o executor valida de novo que o executável está sob a raiz do plugin antes de correr.

## Armadilhas

- Sem **`mb plugins sync`** após alterações manuais na árvore, o Cobra pode não refletir comandos novos.
- **`--no-remove`**: útil para preservar cache, mas `exec_path` pode ficar inválido se o ficheiro foi apagado manualmente.
- Plugins **locais** não são cópia para `PluginsDir` — o path registado tem de continuar válido.

## Helpers de shell (embed)

Ao alterares **helpers** (`*.sh` embebidos) ou a lógica de **embed** / `EnsureShellHelpers`, segue o README do pacote:

- `internal/infra/shellhelpers/README.md` — fluxo (novo `.sh`, `all.sh`, sync, checksum).
- `docs/docs/technical-reference/helpers-shell.md` — referência para utilizadores e lista de helpers.

## Documentação no repositório

- Referência técnica: `docs/docs/technical-reference/plugins.md`, `docs/docs/technical-reference/plugin-invocation-context.md` (`MB_CTX_*`, incluindo `MB_CTX_CHILD_COMMANDS`, `MB_CTX_HIDDEN_CHILD_COMMANDS`, `MB_CTX_CHILD_COMMAND_ALIASES`)
- Criar plugin: `docs/docs/guide/creating-plugins.md`
- Comandos e flags: `docs/docs/guide/plugin-commands.md`
- Helpers de shell: `docs/docs/technical-reference/helpers-shell.md` (alinhado com `internal/infra/shellhelpers/README.md` quando mudas embed); variáveis `MB_CTX_*` em `plugin-invocation-context.md`
- Mapa de pacotes: `internal/README.md`

## Verificação

```bash
go test ./internal/app/plugins/... ./internal/cli/plugins/... ./internal/cli/plugincmd/... ./internal/infra/plugins/... ./internal/domain/plugin/... -count=1
```

Ao alterar comportamento visível ou regras de manifest, alinhar **docs** e mensagens Cobra.
