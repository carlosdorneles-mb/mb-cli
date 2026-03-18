# Pacotes `internal/`

Mapa rápido para onde colocar código novo.

## Composição e CLI


| Pacote                              | Responsabilidade                                                                                                                                    |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `app`                               | Bootstrap Uber Fx (`Bootstrap`), módulos Fx (`PathsModule`, `CacheModule`, …), `NewRuntimeConfig`.                                                  |
| `deps`                              | Paths padrão (`NewPaths`), `RuntimeConfig` (paths + flags), `Dependencies` injetados nos comandos, `LoadDefaultEnvValues` / `SaveDefaultEnvValues`. |
| `commands`                          | Raiz Cobra (`NewRootCmd`), completion tests.                                                                                                        |
| `commands/plugins`, `commands/self` | Subcomandos `mb plugins` e `mb self`.                                                                                                               |
| `plugincmd`                         | Comandos dinâmicos a partir do cache (`plugincmd.Attach`).                                                                                          |


## Domínio


| Pacote       | Responsabilidade                       |
| ------------ | -------------------------------------- |
| `cache`      | SQLite (plugins, categorias, fontes).  |
| `plugins`    | Manifest, scanner de disco, clone Git. |
| `executor`   | Execução segura de scripts de plugin.  |
| `env`        | Merge de variáveis de ambiente.        |
| `selfupdate` | Atualização binária.                   |


## Terminal / TUI

Saída no terminal está dividida em dois pacotes (sem pasta `terminal/` física, para evitar refatoração massiva de imports):

- `**ui**` — Temas (Fang/Gum), banner, mensagens de erro/sucesso em PT, integração leve com Gum (ex. `PrependGumThemeDefaults`).
- `**system**` — Integrações mais pesadas com binários externos: Gum (inputs), Glamour (render de Markdown no README dos plugins).

Regra prática: preferir `ui` para tudo que é “aparência e cópia”; usar `system` quando for necessário chamar Gum/Glamour ou renderizar Markdown de arquivo.

## Utilitários


| Pacote          | Responsabilidade                            |
| --------------- | ------------------------------------------- |
| `safepath`      | Validação de paths sob diretório permitido. |
| `helpers/shell` | Helpers para scripts shell.                 |
| `version`       | Versão de build.                            |


