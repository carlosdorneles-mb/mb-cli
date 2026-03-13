---
sidebar_position: 5
---

# Referência de comandos

## Comandos principais

| Comando | Descrição |
|--------|-----------|
| `mb self sync` | Escaneia o diretório de plugins e atualiza o cache SQLite |
| `mb plugins add <git-url> [--name N] [--tag TAG]` | Instala um plugin a partir de uma URL Git |
| `mb plugins list [--check-updates]` | Lista plugins instalados (name, command, version, url) |
| `mb plugins remove <name>` | Remove um plugin instalado (com confirmação) |
| `mb plugins update [name \| --all]` | Atualiza um plugin ou todos |
| `mb self env list` | Lista variáveis padrão |
| `mb self env set KEY [VALUE]` | Define variável padrão |
| `mb self env unset KEY` | Remove variável padrão |
| `mb <categoria> <comando> [args...]` | Executa o plugin correspondente |

## Completion de shell

O CLI gera scripts de completion para bash, zsh, fish e powershell via `mb completion <shell>`. O completion inclui os comandos built-in (por exemplo `self`, `help`, `completion`) e também **todos os comandos e subcomandos de plugins** disponíveis no cache. Ou seja, após `mb self sync`, ao usar TAB no shell serão sugeridas as categorias e comandos de plugins (ex.: `tools`, `infra`, `tools hello`, `infra ci`).

Para instalar o completion no seu shell, consulte a saída de `mb completion --help` e os subcomandos `bash`, `zsh`, `fish`, `powershell`.

## Flags globais

- `--verbose` — saída mais verbosa
- `--quiet` — reduz mensagens
- `--env-file <path>` — arquivo de variáveis de ambiente
- `--env KEY=VALUE` — injeta variável no processo do plugin (pode ser repetido)

## Testar o CLI

```bash
make test       # testes unitários
make build && ./bin/mb self sync && ./bin/mb plugins list
```

Para testar sem alterar seu config real, use um diretório temporário:

```bash
export XDG_CONFIG_HOME=/tmp/mb-test   # Linux
mkdir -p "$XDG_CONFIG_HOME/mb/plugins/hello"
# ... criar manifest.yaml e run.sh ...
./bin/mb self sync
./bin/mb plugins list
./bin/mb <categoria> hello
```
