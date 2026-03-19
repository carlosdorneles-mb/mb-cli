---
sidebar_position: 3
---

# Referência de comandos

## Comandos principais

| Comando | Descrição |
|--------|-----------|
| `mb update [--only-plugins \| --only-cli]` | Atualiza plugins e o MB CLI. Sem flags: primeiro plugins (como `mb plugins update --all`), depois o binário (como `mb self update`). **`--only-plugins`** só atualiza plugins; **`--only-cli`** só o binário. Não use as duas flags em simultâneo. |
| `mb plugins sync` | Rescaneia o diretório de plugins e paths locais, atualiza o cache SQLite e garante os helpers de shell em `~/.config/mb/lib/shell` |
| `mb self update` | **Só para binários da release oficial** (versão embutida via ldflags no GitHub Release). Builds locais ou `go install` mostram mensagem a usar `install.sh`. Se a release for mais nova, baixa o `mb`, valida SHA256 e substitui o executável (Linux/macOS, amd64/arm64). |
| `mb self update --check-only` | Igual: só em binários de release. Compara com a última release (sem download). **Códigos de saída:** `0` = já atualizado ou versão local mais nova; `2` = há atualização; `1` = erro. Em build local: mensagem + saída `0`. |
| `mb plugins add <url \| path \| .> [--name N] [--tag TAG]` | Instala um plugin: **URL Git** = remoto (clone); **path** ou **`.`** = local (registra o path, sem cópia). **`--name`** = id. da instalação (list/remove/clone), não muda o path do comando. `--tag` só para remoto. |
| `mb plugins list [--check-updates]` | Lista plugins instalados (nome, comando, descrição, versão, **ORIGEM** (local/remoto), URL/path) |
| `mb plugins remove <name>` | Remove um plugin instalado (com confirmação). Se for local, só remove o registro. O cache é atualizado e o plugin deixa de aparecer em `plugins list`. |
| `mb plugins update [name \| --all]` | Atualiza um plugin remoto ou todos (plugins locais não são atualizados) |
| `mb self env list [--group G]` | Lista variáveis em tabela (VAR, GRUPO); com `--group`, só o arquivo `.env.G` |
| `mb self env set <KEY> <VALUE> [--group G]` | Define em `env.defaults` ou em `.env.G` |
| `mb self env unset <KEY> [--group G]` | Remove de `env.defaults` ou de `.env.G` |
| `mb <categoria> <comando> [args...]` | Executa o plugin correspondente (veja [Comandos de plugins](../guide/plugin-commands.md)) |

## Completion de shell

O CLI gera scripts de completion para bash, zsh, fish e powershell via `mb self completion <shell>`. O completion inclui os comandos built-in e **todos os comandos e subcomandos de plugins** disponíveis no cache. Após `mb plugins sync` (ou após instalar um plugin), use TAB para sugerir categorias e comandos de plugins.

Para instalar o completion no seu shell, consulte `mb self completion --help` e os subcomandos `bash`, `zsh`, `fish`, `powershell`.

## Flags globais

- **`--verbose` / `-v`** — Saída mais verbosa. Veja [Flags globais](../guide/global-flags.md).
- **`--quiet` / `-q`** — Reduz mensagens. Veja [Flags globais](../guide/global-flags.md).
- **`--env-file <path>`** — Arquivo de variáveis de ambiente. Veja [Variáveis de ambiente](../guide/environment-variables.md).
- **`--env KEY=VALUE`** — Injeta variável no processo do plugin (pode ser repetido). Veja [Variáveis de ambiente](../guide/environment-variables.md).
- **`--env-group <nome>`** — Sobrepõe `env.defaults` com `~/.config/mb/.env.<nome>` ao executar plugins. Veja [Variáveis de ambiente](../guide/environment-variables.md).
- **`--doc`** — Abre a URL de documentação no navegador (por omissão o site do projeto; configurável em `~/.config/mb/config.yaml` como `docs_url`). Apenas com `mb --doc`, sem subcomando. Veja [Configuração do CLI](cli-config.md) e [Flags globais](../guide/global-flags.md).

## Testar o CLI

```bash
make test       # testes unitários
make build && ./bin/mb plugins sync && ./bin/mb plugins list
```

Para testar sem alterar seu config real, use um diretório temporário:

```bash
export XDG_CONFIG_HOME=/tmp/mb-test   # Linux
mkdir -p "$XDG_CONFIG_HOME/mb/plugins/hello"
# ... criar manifest.yaml e run.sh ...
./bin/mb plugins sync
./bin/mb plugins list
./bin/mb <categoria> hello
```
