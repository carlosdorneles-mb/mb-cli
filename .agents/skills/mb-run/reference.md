# Referência — `mb run`

## `internal/cli/run`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `run.go` | `NewRunCmd` — `DisableFlagParsing`, `runtimeflags.ParseLeadingRuntimeFlags` + `env.ParseInlinePairs`, `MinimumNArgs(1)`, `BuildMergedOSEnviron(d, nil)`, `exec.LookPath`, `exec.CommandContext`, stdio herdados, `os.Exit` no exit code do filho, `SetHelpFunc` delegando no root |

## `internal/cli/runtimeflags`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `runtimeflags.go` | `RegisterRuntimePersistentFlags` (root + mesma semântica), `ParseLeadingRuntimeFlags` (prefixo após `run`) |
| `runtimeflags_test.go` | Casos de peel e merge |
| `run_test.go` | Args obrigatórios, `true`, comando inexistente, injeção de `./.env` no cwd |

## `internal/deps`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `execenv.go` | `BuildMergedOSEnviron`, `BuildMergedOSEnvironWithExtraInline` — `overlay` usado por plugins para `env_files`; **`nil` para `mb run`**; `AppendShellHelpersEnv` → `MB_HELPERS_PATH` (`internal/infra/shellhelpers`) |
| `envdefaults.go` | `BuildEnvFileValues` — `env.defaults`, `--env-group`, `./.env`, `--env-file`; secrets e `op://` |

## Contraste com plugins

| Aspeto | `mb run` | Plugin (`cli/plugincmd`) |
|--------|----------|---------------------------|
| Overlay manifest | **Não** (`overlay == nil`) | Sim — merge de `env_files` do manifest para o grupo efetivo |
| Entrypoint | Executável arbitrário no PATH | Script/binário sob diretório do plugin |

## `internal/cli/root`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `command.go` | `RegisterRuntimePersistentFlags` nas `PersistentFlags`; regista `run.NewRunCmd(d)` com `GroupID` de comandos |
