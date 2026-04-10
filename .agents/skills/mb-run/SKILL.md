---
name: mb-run
description: >-
  Covers MB CLI `mb run`, which executes arbitrary programs with the same merged
  environment as plugins (minus manifest env_files), including MB_HELPERS_PATH. Use when
  changing or explaining `mb run`, subprocess env injection, `BuildMergedOSEnviron`,
  exit codes, PluginTimeout, or docs on run vs plugins; for shell helpers embed see
  mb-plugins skill and `internal/infra/shellhelpers/README.md`.
---

# MB CLI — `mb run`

## Quando aplicar

- Implementar ou corrigir **`mb run`** ou o fluxo de **ambiente** partilhado com plugins.
- Explicar diferenças entre **`mb run`** e execução de **plugin** (sem `env_files` do manifest).
- Ajustar **`DisableFlagParsing`**, **`runtimeflags.ParseLeadingRuntimeFlags`**, ajuda (`mb help run`), timeout, ou propagação de código de saída.

## Comportamento (superfície)

| Aspeto | Detalhe |
|--------|---------|
| Uso | `mb run <comando> [args...]` — o primeiro argumento é resolvido com **`exec.LookPath`** (PATH ou caminho absoluto) |
| Ambiente | **`deps.BuildMergedOSEnviron(d, nil)`** — igual aos plugins para camadas de ficheiro + inline, **sem** overlay de `env_files` do manifest (`overlay == nil`) |
| Flags globais do `mb` | **Antes** de `run`, ou **prefixo logo após** `run` (antes do executável): `-e`/`--env`, `--env-file`, `--env-vault`, `-v`/`--verbose`, `-q`/`--quiet` — via **`runtimeflags.ParseLeadingRuntimeFlags`** em conjunto com o que o root já parseou. Com **`DisableFlagParsing: true`**, o Cobra **não** parseia; o `run` descasca só esse prefixo e o resto vai ao filho (ex.: `mb run grep -r`) |
| Ajuda | **`mb help run`** — `mb run --help` pode ser entregue ao executável filho (documentado no Long) |
| Timeout | Se **`Runtime.PluginTimeout > 0`**, o contexto do **`exec.CommandContext`** tem deadline (mesma config que plugins) |
| Código de saída | Em **`exec.ExitError`**, o processo termina com **`os.Exit(código)`** para propagar o exit code do filho |
| Erros | Comando não encontrado no PATH → erro Go (não `os.Exit`) |

Para a **ordem completa** de variáveis (`env.defaults`, `--env-vault`, `./.env`, `--env-file`, `--env`, secrets, `*.opsecrets`, `op://`, tema gum, `MB_VERBOSE`/`MB_QUIET`, `MB_HELPERS_PATH`), ver a skill **`mb-envs`** e `docs/docs/guide/environment-variables.md`.

## Onde está o código

| Área | Caminho |
|------|---------|
| Comando | `internal/cli/run/run.go` — `NewRunCmd` |
| Flags globais partilhadas | `internal/cli/runtimeflags/runtimeflags.go` — `RegisterRuntimePersistentFlags`, `ParseLeadingRuntimeFlags` |
| Root | `internal/cli/root/command.go` — chama `RegisterRuntimePersistentFlags` nas `PersistentFlags` |
| Merge de ambiente | `internal/deps/execenv.go` — `BuildMergedOSEnviron`, `BuildMergedOSEnvironWithExtraInline` |
| Camadas de ficheiro / secrets | `internal/deps/envdefaults.go` — `BuildEnvFileValues` |

Detalhe por ficheiro: [reference.md](reference.md).

## Regras de produto (resumo)

1. **`mb run` não lê `env_files` do manifest** — só plugins aplicam esse overlay (`plugincmd` passa overlay não-nulo). O utilizador usa **`--env-file`** na linha do `mb` se precisar de ficheiros extra.
2. O ambiente efetivo é o mesmo **pipeline** que plugins: sistema → ficheiros mb → `--env` com maior precedência, mais injeções fixas (gum, verbosidade, helpers).
3. **Stdin/stdout/stderr** são os do terminal (`os.Stdin`, etc.).

## Armadilhas

- Esquecer que só o **prefixo** após `run` é MB — **`mb run cmd --env X=1`** envia `--env` ao **filho**; usar **`mb run -e X=1 cmd`** ou **`mb --env X=1 run cmd`**.
- Confundir **`mb run myplugin`** com **`mb categoria comando`** — `run` é só executável no PATH, não resolve comandos do cache SQLite.

## Documentação no repositório

- `docs/docs/guide/environment-variables.md` — precedência e `mb run`
- `docs/docs/technical-reference/reference.md` — linha `mb run`
- **Helpers de shell** — o ambiente inclui `MB_HELPERS_PATH` (mesmo pipeline que plugins). Se alterares embed ou scripts helpers, segue `internal/infra/shellhelpers/README.md` e atualiza `docs/docs/technical-reference/helpers-shell.md` (ver skill **`mb-plugins`**).

## Verificação

```bash
go test ./internal/cli/run/... ./internal/cli/runtimeflags/... ./internal/cli/root/... ./internal/deps/... -count=1
```

Ao alterar merge de ambiente partilhado, validar também **`cli/plugincmd`** e testes em **`internal/deps`**.
