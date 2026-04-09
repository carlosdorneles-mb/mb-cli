---
sidebar_position: 3
---

# Contexto de invocação de plugins (`MB_CTX_*`)

Quando o MB **executa um plugin** (script ou binário registado no manifest), o processo do plugin recebe variáveis de ambiente adicionais além de `MB_HELPERS_PATH`, `MB_QUIET` e `MB_VERBOSE`. Todas seguem o prefixo **`MB_CTX_`** para se distinguirem das restantes.

Servem para o script saber **como** o utilizador chamou o `mb` (linha de comandos), **qual** comando do manifest corresponde à folha atual, **que** flags do plugin foram usadas e **que** outros subcomandos existem ao mesmo nível na árvore do CLI.

Esta página é a **referência das variáveis** definidas pelo runtime do CLI. Para funções shell que apenas **lêem** essas variáveis (`mb_context_dump`, `mb_peer_commands_lines`), veja o helper **context** em [Helpers de shell](helpers-shell.md#context).

## Variáveis

| Variável | Descrição |
|----------|-----------|
| `MB_CTX_INVOCATION` | Linha de argumentos vista pelo processo do `mb`: junção de `argv` com espaços (o primeiro elemento pode ser o caminho absoluto ao binário). |
| `MB_CTX_CONFIG_DIR` | Diretório de configuração do MB (por exemplo `~/.config/mb`), útil para localizar `cache.db` ou outros ficheiros. |
| `MB_CTX_COMMAND_PATH` | Caminho lógico do comando no manifest / cache (ex.: `tools`, `tools/vscode`). |
| `MB_CTX_COMMAND_NAME` | Último segmento de `MB_CTX_COMMAND_PATH` (ex.: `vscode`). |
| `MB_CTX_PARENT_COMMAND_PATH` | Caminho do “pai” no manifest (tudo antes do último `/`); vazio se a folha está na raiz (ex.: só `hello`). |
| `MB_CTX_COBR_COMMAND_PATH` | Caminho Cobra (`cmd.CommandPath()`), normalmente `mb` + subcomandos (ex.: `mb tools vscode`); pode diferir de `MB_CTX_COMMAND_PATH` quando há aliases. |
| `MB_CTX_PLUGIN_FLAGS` | Nomes **longos** das flags do plugin que foram passadas (separados por espaço), ordenados. Em comandos só com entrypoint, inclui flags locais alteradas exceto `--readme`. Em comandos com `flags` no manifest, só entram flags definidas no manifest. |
| `MB_CTX_PEER_COMMANDS` | JSON com um array de strings: nomes dos **outros** comandos irmãos sob o mesmo pai Cobra (ex.: outras folhas sob `mb tools`), ordenados; o comando atual não entra na lista. |

## Exemplo: comando aninhado com flag

Se o manifest define o comando em `tools/vscode` e o utilizador executa (valores ilustrativos):

```bash
mb tools vscode --install
```

o ambiente do script do plugin pode conter algo deste género:

```text
MB_CTX_INVOCATION=/usr/local/bin/mb tools vscode --install
MB_CTX_CONFIG_DIR=/home/user/.config/mb
MB_CTX_COMMAND_PATH=tools/vscode
MB_CTX_COMMAND_NAME=vscode
MB_CTX_PARENT_COMMAND_PATH=tools
MB_CTX_COBR_COMMAND_PATH=mb tools vscode
MB_CTX_PLUGIN_FLAGS=install
MB_CTX_PEER_COMMANDS=["bruno","flutter"]
```

`MB_CTX_PEER_COMMANDS` lista os **outros** nomes de comando irmãos (por exemplo `bruno`, `flutter`), não inclui `vscode`.

## Exemplo: folha na raiz do manifest

Para um comando registado só como `hello` e invocação `mb hello -q` (o `-q` é flag global do `mb`, não do plugin):

```text
MB_CTX_INVOCATION=/usr/local/bin/mb hello -q
MB_CTX_COMMAND_PATH=hello
MB_CTX_COMMAND_NAME=hello
MB_CTX_PARENT_COMMAND_PATH=
MB_CTX_COBR_COMMAND_PATH=mb hello
MB_CTX_PLUGIN_FLAGS=
MB_CTX_PEER_COMMANDS=["deploy","status"]
```

(Valores ilustrativos: são os outros comandos irmãos sob o mesmo pai que `hello`.)

`MB_CTX_PLUGIN_FLAGS` fica vazio se nenhuma flag **do plugin** tiver sido passada; `MB_QUIET` continua a refletir `-q` via `MB_QUIET=1`.

## Uso direto no script

```sh
#!/usr/bin/env bash
. "$MB_HELPERS_PATH/all.sh"

log info "Comando manifest: ${MB_CTX_COMMAND_PATH:-?}"
log info "Invocação Cobra: ${MB_CTX_COBR_COMMAND_PATH:-?}"

if [[ -n "${MB_CTX_PLUGIN_FLAGS:-}" ]]; then
  log debug "Flags do plugin: $MB_CTX_PLUGIN_FLAGS"
fi

# Cache SQLite (cuidado: o schema pode mudar entre versões do mb)
# cache_db="${MB_CTX_CONFIG_DIR}/cache.db"
```

## Limitação

Ao executar apenas `mb <categoria>` sem escolher uma folha (só ajuda do grupo), **não** corre nenhum plugin — portanto essas variáveis não são definidas nesse caso. Para listar filhos nessa situação, use `mb help <categoria>`, completion, ou consulte o cache com cuidado (schema pode evoluir).

## Privacidade

`MB_CTX_INVOCATION` pode conter segredos se o utilizador os passar na linha de comando; evite registar ou enviar este valor em claro.

## Onde isto é definido no código

As variáveis são montadas na execução de plugins (`internal/cli/plugincmd`, função `appendPluginInvocationEnv` em `context_env.go`).

## Ver também

- [Plugins](plugins.md) — cache, manifest e execução
- [Helpers de shell — context](helpers-shell.md#context) — helper `context.sh`: `mb_context_dump`, `mb_context_dump_json`, `mb_peer_commands_lines`, `mb_ctx_has_plugin_flag`, `mb_ctx_peer_contains`, `mb_ctx_peer_count`, `mb_ctx_cache_db`, `mb_ctx_parent_is`, `mb_ctx_command_path_is`, `mb_ctx_path_depth`
