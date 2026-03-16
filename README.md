# MB CLI

CLI em Go para orquestrar plugins com descoberta dinâmica (cache SQLite) e injeção de variáveis de ambiente.

<img title="MB CLI" alt="Alt text" src="mb-cli.png">

## Instalação

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash
```

Requer `~/.local/bin` no PATH. Depois: `mb self sync`.

Versão específica: `bash -s -- --version 0.0.5`. Desinstalar: `curl -sSL .../uninstall.sh | bash`.

## Build (desenvolvedores)

```bash
make build    # bin/mb
make install  # $GOPATH/bin
```

## Release

SemVer via tags Git. `make deps` e `make release` (menu 1–3) ou workflow **Bump version** no GitHub Actions. Detalhes: [documentação](https://carlosdorneles-mb.github.io/mb-cli/docs/versioning-and-release).

## Documentação

A documentação está em **https://carlosdorneles-mb.github.io/mb-cli/**.

- [Começar](https://carlosdorneles-mb.github.io/mb-cli/docs/getting-started) — pré-requisitos, instalação
- [Comandos e plugins](https://carlosdorneles-mb.github.io/mb-cli/docs/plugin-commands)
- [Criar um plugin](https://carlosdorneles-mb.github.io/mb-cli/docs/creating-plugins)
- [Referência](https://carlosdorneles-mb.github.io/mb-cli/docs/reference)

## Comandos

| Comando | Descrição |
|--------|-----------|
| `mb self sync` | Atualiza cache de plugins |
| `mb plugins add <url \| path>` | Instala plugin |
| `mb plugins list` | Lista plugins |
| `mb <categoria> <comando>` | Executa plugin |

Flags: `-v` / `--verbose`, `-q` / `--quiet`, `--env KEY=VALUE`, `--env-file <path>`.
