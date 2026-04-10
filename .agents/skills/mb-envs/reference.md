# Referência — ficheiros `mb envs` e ambiente

## `internal/usecase/envs`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `set.go` | `Set(...)` — plain, `--secret`, `--secret-op` (`*.opsecrets`, sem keyring para `op://` novos) |
| `unset.go` | `Unset(...) (removed bool, err)` — ficheiro, `.secrets`, `.opsecrets`, keyring, 1Password |
| `list.go` | `CollectListRows`, `CountListableEnvKeys`, `rowsForPath`; `--vault project` / `project/x` só YAML (`project/x` = só sub-mapa `envs.x`) |
| `vaults.go` | `CollectVaultRows` — `default`, `project` / `project/*` (`mbcli.yaml`), glob `.env.<vault>` (ignora reservado `project`, sidecars) |
| `paths.go` | `TargetPath` (`ValidateConfigurableVault`), `KeyringGroup` |
| `secretpref.go` | `ResolveSetSecretFlags`, `MB_ENVS_SECRET_STORE` |

## `internal/cli/envs`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `env.go` | Agrega subcomandos `list`, `vaults`, `set`, `unset` |
| `set.go` | `KEY=VALOR` múltiplos; `--vault`, `--secret`, `--secret-op`, `--yes` |
| `unset.go` | Várias chaves; `--vault` |
| `list.go` | Tabela VAR, VAULT, ARMAZENAMENTO; `--json`, `--text`, `--show-secrets`, `--vault` |
| `vaults.go` | Tabela VAULT, ARQUIVO; `--json` / `-J` |
| `path.go` | `envPaths(d)` → `appenvs.Paths` |

## `internal/deps`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `envdefaults.go` | `LoadDefaultEnvValues`, `BuildEnvFileValues`, `mergeSecretKeysInto`, `mergeOPRefsInto`, `resolveSecretValueForMerge` |
| `execenv.go` | `BuildMergedOSEnviron` — orquestra ficheiros + `--env` + tema gum + `MB_*` |
| `secretkeys.go` | `LoadSecretKeys`, `AddSecretKey`, `RemoveSecretKey` (`path + ".secrets"`) |
| `opsecrets.go` | `LoadOPSecretRefs`, `SetOPSecretRef`, `RemoveOPSecretRef` (`path + ".opsecrets"`) |
| `deps.go` | `Dependencies` com `SecretStore`, `OnePassword`; `RuntimeConfig.EnvVault` |

## `internal/shared/envvault`

Validação de nomes de vault e `FilePath` para `.env.<vault>`.

## `internal/infra/opcli`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `client.go` | `PutSecret`, `RemoveSecretField`, `ReadOPReference`; item title `mb-cli env / <vault>` |

## Manifest de plugins

`internal/infra/plugins/manifest_env.go` — `env_files` por **vault** (campo YAML/JSON `vault`).
