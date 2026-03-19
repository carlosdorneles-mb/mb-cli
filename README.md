# MB CLI

CLI para orquestrar plugins com descoberta dinâmica e injeção de variáveis de ambiente. **Implementação:** Go; árvore de comandos (Cobra), cache local (SQLite), UI no terminal (Charm / Fang). Documentação de uso: site abaixo; detalhes técnicos em [Arquitetura](https://carlosdorneles-mb.github.io/mb-cli/docs/architecture).

<img title="MB CLI" alt="Alt text" src="mb-cli.png">

## Instalação

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash
```

Requer `~/.local/bin` no PATH. Depois: `mb plugins sync`.

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
| `mb update [--only-plugins \| --only-cli \| --only-system]` | Atualiza plugins, binário do MB e sistema (brew/mas ou apt/flatpak/snap); sem flags executa as três fases; `--only-*` combináveis; `--check-only` só com `--only-cli` |
| `mb plugins sync [--no-remove]` | Atualiza cache; opcionalmente mantém comandos órfãos |
| `mb update --only-cli` | Atualiza o binário `mb` (só binários da release oficial) |
| `mb update --only-cli --check-only` | Verifica atualização (release); saída `2` se houver |
| `mb plugins add <url \| path>` | Instala ou substitui pacote; `--no-remove` repassa ao sync |
| `mb plugins list` | Lista plugins |
| `mb <categoria> <comando>` | Executa plugin |

Flags: `-v` / `--verbose`, `-q` / `--quiet`, `--env KEY=VALUE`, `--env-file <path>`, `--doc` (abre a documentação; URL em `~/.config/mb/config.yaml`, ver [referência técnica](https://carlosdorneles-mb.github.io/mb-cli/docs/technical-reference/cli-config)).
