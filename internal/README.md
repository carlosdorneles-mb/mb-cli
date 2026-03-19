# Pacotes `internal/`

Mapa rápido para onde colocar código novo. Estrutura orientada a FX (bootstrap → module → cli; shared, domain, ports, infra, app).

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
| App     | `app/plugins`                          | Use cases: RunSync (sync), add/remove/update (lógica aplicação).                  |
| Infra   | `infra/sqlite`                         | Store SQLite (plugins, categorias, fontes).                                       |
|         | `infra/plugins`                        | Manifest, scanner de disco, clone Git.                                           |
|         | `infra/executor`                       | Execução segura de scripts de plugin.                                            |
|         | `infra/browser`, `infra/selfupdate`    | Abertura de URL, atualização binária.                                            |
|         | `infra/shellhelpers`                   | Helpers para scripts shell (embed `.sh`, EnsureShellHelpers).                     |
| Shared  | `shared/ui`, `shared/system`           | Temas (Fang/Gum), banner, mensagens; Gum/Glamour, Markdown.                       |
|         | `shared/safepath`, `shared/version`    | Validação de paths, versão de build.                                             |
|         | `shared/env`, `shared/envgroup`       | Merge de variáveis de ambiente, grupos.                                          |
|         | `shared/config`                        | AppConfig, Load, DefaultDocsURL.                                                  |
| Domain  | `domain/plugin`                        | Tipos/regras de domínio (placeholder; opcional).                                 |
| Deps    | `deps`                                 | Paths padrão, RuntimeConfig, Dependencies (injetados nos comandos até migração). |

## Ordem de dependência (evitar ciclos)

`shared` → `domain`/`ports` → `infra` → `app` → `module` → `cli` → `bootstrap`.

## Terminal / TUI

- **`shared/ui`** — Temas (Fang/Gum), banner, mensagens de erro/sucesso em PT.
- **`shared/system`** — Gum (inputs), Glamour (render de Markdown no README dos plugins).

Regra: preferir `ui` para aparência e cópia; usar `system` quando for necessário chamar Gum/Glamour ou renderizar Markdown de arquivo.
