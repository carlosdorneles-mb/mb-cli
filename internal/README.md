# Pacotes `internal/`

Mapa rápido para onde colocar código novo. Estrutura orientada a FX (bootstrap → module → cli; domain, ports, app, infra, shared).

## Mapa atual (pós-reorganização)

| Camada   | Pacote / path                         | Responsabilidade                                                                 |
| -------- | ------------------------------------- | --------------------------------------------------------------------------------- |
| Bootstrap | `bootstrap`                           | `fx.New`, Options agregando todos os módulos, `Populate(&rootCmd)`.               |
| Module  | `module/runtime`                      | Paths, RuntimeConfig, NewAppConfig (providers FX).                               |
|         | `module/cache`                        | newStore, registerStoreLifecycle (SQLite).                                        |
|         | `module/plugins`, `module/executor`   | Scanner, Executor (providers FX).                                                |
|         | `module/deps`                         | DepsModule: RuntimeConfig, AppConfig, Store, Scanner, Executor.                   |
|         | `module/cli`                           | CLIModule: Provide(root.NewRootCmd).                                             |
| CLI     | `cli/root`                             | Raiz Cobra (`NewRootCmd`), completion tests.                                     |
|         | `cli/plugins`, `cli/envs`, `cli/update` | Subcomandos `mb plugins`, `mb envs`, `mb update` (e `mb completion` no root).  |
|         | `cli/plugincmd`                        | Comandos dinâmicos a partir do cache (`Attach`).                                 |
| App     | `app/plugins`                          | Casos de uso (ex.: `RunSync`) via **`ports`** + **`domain/plugin`** (sem importar `infra`). |
| Ports   | `ports`                                | Interfaces estáveis (ex.: `PluginSyncStore`, `PluginScanner`, `ShellHelperInstaller`). |
| Domain  | `domain/plugin`                        | DTOs do cache (`Plugin`, `Category`, …), `ValidationWarning`, `HelpGroupDef`, merge/validação de groups. |
| Infra   | `infra/sqlite`                         | Store SQLite; tipos expostos como alias dos do domínio (`type Plugin = plugin.Plugin`). |
|         | `infra/plugins`                        | Manifest, scanner de disco, clone Git; implementa `ports.PluginScanner`.           |
|         | `infra/executor`                       | Execução segura de scripts de plugin.                                            |
|         | `infra/browser`, `infra/selfupdate`    | Abertura de URL, atualização binária.                                            |
|         | `infra/shellhelpers`                   | Helpers shell embed + `Installer` que implementa `ports.ShellHelperInstaller`.   |
| Shared  | `shared/ui`, `shared/system`           | Temas (Fang/Gum), banner, mensagens; Gum/Glamour, Markdown.                       |
|         | `shared/safepath`, `shared/version`    | Validação de paths, versão de build.                                             |
|         | `shared/env`, `shared/envgroup`       | Merge de variáveis de ambiente, grupos.                                          |
|         | `shared/config`                        | AppConfig, Load, DefaultDocsURL.                                                  |
| Deps    | `deps`                                 | Paths padrão, RuntimeConfig, `Dependencies` (bundle FX para comandos).            |

## Ordem de dependência (evitar ciclos)

- **`domain/plugin`**: só stdlib + libs puras (ex.: YAML para parse de groups).
- **`ports`**: depende de `domain/plugin` (tipos nos contratos).
- **`app/*`**: depende de `domain`, `ports`, `shared` — **não** de `infra/*`.
- **`infra/*`**: implementa `ports` e usa `domain` + `sqlite`/FS/rede.
- **`cli/*`**: Cobra, chama `app` via adaptadores (ex.: `cli/plugins/sync_run.go` injeta `Store`, `Scanner`, `shellhelpers.Installer`).
- **`module/*`**, **`bootstrap`**: composição FX.

## Terminal / TUI

- **`shared/ui`** — Temas (Fang/Gum), banner, mensagens de erro/sucesso em PT.
- **`shared/system`** — Gum (inputs), Glamour (render de Markdown no README dos plugins).

Regra: preferir `ui` para aparência e cópia; usar `system` quando for necessário chamar Gum/Glamour ou renderizar Markdown de arquivo.
