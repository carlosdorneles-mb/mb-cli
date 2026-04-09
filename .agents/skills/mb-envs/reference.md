# Referência — ficheiros `mb envs` e ambiente

## `internal/app/envs`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `set.go` | `Set(...)` — plain, `--secret`, `--secret-op`; transição plain↔secret remove keyring/`op` quando aplicável |
| `unset.go` | `Unset(...) (removed bool, err)` — só grava se chave existia (ficheiro ou `.secrets`) |
| `list.go` | `CollectListRows`, `rowsForPath`, `secretStorageFromStored` (`local` / `keyring` / `1password`), `resolveStoredSecretForList` (`op://` → `ReadOPReference`) |
| `groups.go` | `CollectEnvGroupRows` — `default` + `env.defaults`, glob `.env.<grupo>` (ignora `*.secrets`) |
| `paths.go` | `TargetPath`, `KeyringGroup` |

## `internal/cli/envs`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `env.go` | Agrega subcomandos `list`, `set`, `unset` |
| `set.go` | Flags `--group`, `--secret`, `--secret-op` (mutually exclusive) |
| `unset.go` | Mensagens para `removed` vs inexistente |
| `list.go` | Tabela VAR, GRUPO, ARMAZENAMENTO; `--json`, `--text`, `--show-secrets`, `--group` |
| `groups.go` | Tabela GRUPO, ARQUIVO; `--json` / `-J`; alias `group` |
| `path.go` | `envPaths(d)` → `appenvs.Paths` |

## `internal/deps`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `envdefaults.go` | `LoadDefaultEnvValues`, `BuildEnvFileValues`, `mergeSecretKeysInto`, `resolveSecretValueForMerge` (`op://`) |
| `execenv.go` | `BuildMergedOSEnviron` — orquestra ficheiros + `--env` + tema gum + `MB_*` |
| `secretkeys.go` | `LoadSecretKeys`, `AddSecretKey`, `RemoveSecretKey` (`path + ".secrets"`) |
| `deps.go` | `Dependencies` com `SecretStore`, `OnePassword` |

## `internal/infra/opcli`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `client.go` | `PutSecret`, `RemoveSecretField`, `ReadOPReference`; `op item get/create/edit`; `isItemNotFound` |
| `itemjson.go` | `upsertConcealedFieldInItemJSON`, `marshalItemJSONForOPEdit` (omitir `vault`), `fieldReferenceFromItemJSON` |

## `internal/ports/onepassword.go`

Interface `OnePasswordEnv` — implementação concreta: `*opcli.Client`.

## Manifest de plugins

`internal/infra/plugins/manifest_env.go` — `env_files` por grupo (overlay após ficheiros globais, antes de `--env` na prática de precedência documentada).
