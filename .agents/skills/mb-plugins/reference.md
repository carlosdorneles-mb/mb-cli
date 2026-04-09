# Referência — `mb plugins` e pacotes relacionados

## `internal/app/plugins`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `sync.go` | `RunSync`, `SyncOptions` (`EmitSuccess`, `NoRemove`, `PostSync`), `SyncReport`, diff added/updated/removed |
| `runtime.go` | `PluginRuntime` — dados passados ao sync (paths, flags, store, scanner, …) |
| `add_remote.go` | `RunAddRemote` — clone Git, registo em fontes, sync |
| `add_local.go` | `RunAddLocalPath` — validação de layout, registo `local_path` |
| `remove.go` | `RunRemovePackage` — remove fonte + árvore em `PluginsDir` quando remoto |
| `update.go` | `RunUpdateAllGitPlugins`, `UpdateOneRemotePackage` — pull/fetch + sync |
| `fsutil.go` | Utilitários de filesystem no fluxo de plugins |

## `internal/cli/plugins`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `plugins.go` | `NewPluginsCmd` — subcomandos add, list, remove, update, sync |
| `sync.go` / `sync_run.go` | Cobra `sync`; `RunSync` delega em `app/plugins.RunSync` |
| `add.go` | Parse URL vs path; `RunAddRemote` / `RunAddLocalPath` |
| `list.go` | Lista + origem; opcional `--check-updates` |
| `remove.go` | Confirmação + `RunRemovePackage` |
| `update.go` | Pacote único ou `--all` |

## `internal/infra/plugins`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `scanner.go` | `Scanner` — descobre `manifest.yaml`, monta `command_path`, tipos sh/bin |
| `manifest.go` | Parse/validação manifest, categorias, flags, entrypoints |
| `manifest_env.go` | `env_files` por grupo, merge para execução |
| `git.go` / `git_service.go` | Operações Git (clone, fetch, tags) |
| `layout_validator.go` | `LayoutValidator` — paths permitidos no pacote |
| `source.go` | `SourceForPlugin`, resolução de diretório de instalação |
| `groups.go` | `groups.yaml` — lotes de help groups |
| `plugin_leaf_hash.go` | Digest por comando (folha) para detetar alterações |

## `internal/cli/plugincmd`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `attach.go` | Anexa árvore de comandos do cache ao root Cobra |
| `run.go` | Executa plugin: env manifest, flags, `ScriptExecutor` |
| `leaf.go` | Folhas e flags-only |

## Outros

| Path | Notas |
|------|--------|
| `internal/domain/plugin/` | DTOs (`Plugin`, categorias, validação) expostos via SQLite |
| `internal/infra/sqlite/` | `ListPlugins`, escrita após sync, `plugin_sources` |
| `internal/infra/executor/` | Execução com verificação de path sob raiz do plugin |
| `internal/infra/shellhelpers/` | `EnsureShellHelpers`, embed `*.sh` → `lib/shell`; `MB_HELPERS_PATH` (ver README do pacote) |
| `internal/ports/` | `PluginScanner`, `PluginSyncStore`, `ScriptExecutor`, `GitOperations`, … |
| `internal/cli/root/` | Registo do grupo `plugins` e `AttachDynamicCommands` após sync |
