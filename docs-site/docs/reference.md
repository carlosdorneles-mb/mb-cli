---
sidebar_position: 5
---

# Referência de comandos

## Comandos principais

| Comando | Descrição |
|--------|-----------|
| `mb self sync` | Escaneia o diretório de plugins e atualiza o cache SQLite |
| `mb self list` | Lista todos os comandos disponíveis (cache) |
| `mb self env list` | Lista variáveis padrão |
| `mb self env set KEY [VALUE]` | Define variável padrão |
| `mb self env unset KEY` | Remove variável padrão |
| `mb <categoria> <comando> [args...]` | Executa o plugin correspondente |

## Flags globais

- `--verbose` — saída mais verbosa
- `--quiet` — reduz mensagens
- `--env-file <path>` — arquivo de variáveis de ambiente
- `--env KEY=VALUE` — injeta variável no processo do plugin (pode ser repetido)

## Testar o CLI

```bash
make test       # testes unitários
make build && ./bin/mb self sync && ./bin/mb self list
```

Para testar sem alterar seu config real, use um diretório temporário:

```bash
export XDG_CONFIG_HOME=/tmp/mb-test   # Linux
mkdir -p "$XDG_CONFIG_HOME/mb/plugins/hello"
# ... criar manifest.yaml e run.sh ...
./bin/mb self sync
./bin/mb self list
./bin/mb <categoria> hello
```
