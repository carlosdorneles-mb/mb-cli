# Referência — `mb update`

## `internal/app/update`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `orchestrate.go` | `Run`, `Options`, `ResolveUpdatePhases`; validação `--check-only` / `--json`; ordem plugins → tools → CLI → sistema |
| `cli.go` | `RunCLIUpdate`, `CLIUpdateNonReleaseMsg`; integração `selfupdate` + `shellhelpers` |
| `tools.go` | `RunToolsUpdateAllPhase`, `findToolsUpdateAllCmd` — nested `tools --update-all` |
| `machine_phase.go` | `RunMachineUpdatePhase`, `findMachineUpdateCmd` — nested `machine update` |
| `helpers.go` | `SelfUpdateConfigFromDeps`, logging auxiliar para check-only e mensagens multi-linha |
| `*_test.go` | Fases, orquestração, CLI, tools, machine |

## `internal/cli/update`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `update.go` | `NewUpdateCmd` — flags `--only-*`, `--check-only`, `--json`; constantes `machineSystemUpdateCommandPath`, `toolsPluginCommandPath`; injeta `RunAllGitPlugins` |

## `internal/cli/plugins`

| Ficheiro | Responsabilidade |
|----------|------------------|
| `update.go` | `RunUpdateAll`, `RunUpdate` por pacote — usado como fase **plugins** do `mb update` |

## `internal/infra/selfupdate`

| Conteúdo típico | Responsabilidade |
|-----------------|------------------|
| Config, `Run`, `RunCheckOnly`, `CheckOnlyDetails` | Download release, SHA256, check-only, códigos de saída (`ExitCodeUpdateAvailable`, etc.) |

## `internal/shared/version`

| Conteúdo | Responsabilidade |
|----------|------------------|
| `Version`, `IsReleaseBuild` | Binário release vs build local — gating do self-update |
