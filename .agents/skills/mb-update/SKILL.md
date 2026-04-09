---
name: mb-update
description: >-
  Covers MB CLI phased updates (`mb update`): Git plugins, `mb tools --update-all`,
  self-update from GitHub Releases, and `mb machine update` via plugin. Use when
  changing or explaining `mb update`, `--only-*`, `--check-only`, `--json`, exit
  codes, `app/update` orchestration, `infra/selfupdate`, or docs on update /
  versioning.
---

# MB CLI — `mb update`

## Quando aplicar

- Implementar ou corrigir **`mb update`** ou a orquestração em **`internal/app/update`**.
- Alterar **self-update** (`infra/selfupdate`), **check-only**, **`--json`**, códigos de saída.
- Ajustar fases **tools** (`mb tools --update-all`) ou **sistema** (`mb machine update` aninhado).
- Atualizar **documentação** (getting-started, reference, cli-config `update_repo`, versioning).

## Comandos e flags (superfície)

| Flag / uso | Notas |
|------------|--------|
| *(nenhum `--only-*`)* | Corre **todas** as fases habilitadas em runtime (plugins → tools → CLI → sistema) |
| `--only-plugins` | Só atualiza plugins remotos (Git) — delega em `plugins.RunUpdateAll` |
| `--only-tools` | Só **`mb tools --update-all`**; a flag **só aparece na ajuda** se o cache tiver o comando `tools` com **`update-all`** no manifest (após `mb plugins sync`) |
| `--only-cli` | Download/substituição do binário **release** (ldflags); builds locais / `go install` → mensagem e **não** atualiza |
| `--only-cli --check-only` | Sem download; compara com a API de releases. **`--json`** só neste modo. Saída **`2`** se houver atualização (scripts) |
| `--only-system` | Executa **`mb machine update`** no root Cobra; flag **só na ajuda** se **`machine/update`** estiver no cache |
| Combinações `--only-*` | Várias podem ser verdadeiras em conjunto; a ordem das fases é sempre a mesma (ver abaixo) |

Restrições validadas em código: **`--check-only`** apenas com **`--only-cli`**; **`--json`** apenas com **`--only-cli --check-only`**.

## Ordem das fases (`app/update.Run`)

1. **Plugins** — `Options.RunAllGitPlugins` (em `cli/update` passa-se `plugins.RunUpdateAll`).
2. **Tools** — `RunToolsUpdateAllPhase`: se existir `tools` com flag **`update-all`**, corre `root.ExecuteContext` com args `tools --update-all`; senão, no-op (aviso se **`--only-tools`** for a única fase).
3. **CLI** — Se `--check-only`: `selfupdate.CheckOnlyDetails` / `RunCheckOnly` (e opcionalmente JSON no stdout). Senão: `RunCLIUpdate` → `selfupdate.Run` para builds release; depois `shellhelpers.EnsureShellHelpers` em erro best-effort.
4. **Sistema** — `RunMachineUpdatePhase`: se existir `machine update` no root, `root.ExecuteContext` com args `machine update`; senão, no-op (aviso se **`--only-system`** for a única fase).

## Onde está o código

| Área | Caminho |
|------|---------|
| Orquestração | `internal/app/update/` (`orchestrate.go` — `Run`, `ResolveUpdatePhases`; `cli.go` — `RunCLIUpdate`; `tools.go` — fase tools; `machine_phase.go` — fase sistema; `helpers.go`) |
| Cobra | `internal/cli/update/update.go` — `NewUpdateCmd`, deteção dinâmica de flags (`storeHasToolsUpdateAll`, `storeHasMachineSystemUpdate`) |
| Atualização de plugins (fase 1) | `internal/cli/plugins/update.go` — `RunUpdateAll` |
| Self-update / releases | `internal/infra/selfupdate/` |
| Versão embutida | `internal/shared/version/` (`IsReleaseBuild`, etc.) |

Detalhe por ficheiro: [reference.md](reference.md).

## Regras de produto (resumo)

1. **CLI release**: `mb update --only-cli` só substitui binário quando **`version.IsReleaseBuild()`**; caso contrário imprime texto a apontar para `install.sh` / releases.
2. **`update_repo`**: configurável (veja `docs/docs/technical-reference/cli-config.md`) para API GitHub Releases.
3. **Fases tools/sistema** dependem de comandos **dinâmicos** registados após sync — sem plugin, a fase não corre (comportamento documentado).

## Armadilhas

- **`--only-tools`** / **`--only-system`** ausentes na ajuda ≠ bug: dependem do **cache SQLite** após **`mb plugins sync`**.
- Fase **sistema** usa **`cmd.Root().ExecuteContext`** com args temporários — alterações no root Cobra podem afetar `findMachineUpdateCmd` / `findToolsUpdateAllCmd`.
- **`check-only`** com código **2**: o processo pode terminar com **`os.Exit(2)`** após imprimir JSON (fluxo em `orchestrate.go`).

## Documentação no repositório

- `docs/docs/guide/getting-started.md` — secção sobre `mb update`
- `docs/docs/technical-reference/reference.md` — tabela de variantes
- `docs/docs/technical-reference/cli-config.md` — `update_repo`
- `docs/docs/technical-reference/versioning-and-release.md`
- `docs/docs/guide/global-flags.md` — exemplo `-q update --only-cli --check-only`

## Verificação

```bash
go test ./internal/app/update/... ./internal/cli/update/... ./internal/infra/selfupdate/... -count=1
```

Ao alterar fases ou flags, atualizar **docs** e a **ajuda longa** em `cli/update/update.go`.
