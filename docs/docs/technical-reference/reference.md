---
sidebar_position: 3
---

# Referência de comandos

## Comandos principais

| Comando | Descrição |
| -------- | ----------- |
| `mb update [--only-plugins \| --only-cli]` | Sem flags: **plugins** → **tools** (`mb tools --update-all`, se o agregador `tools` com essa flag existir no cache) → **MB CLI** → **sistema** via **`mb machine update`** (plugin shell **`machine/update`**: `brew`/`mas` ou `apt`/`flatpak`/`snap`). Sem **`machine/update`** no cache, a fase sistema é ignorada (aviso se **só** `--only-system`). Com **`--only-*`**, escolhe fases; pode **combinar** várias. **`--check-only`** só com **`--only-cli`**. Com **`tools`** + **`--update-all`** no cache, a ajuda inclui **`--only-tools`**. Com **`machine/update`** no cache após **`mb plugins sync`**, inclui **`--only-system`**. |
| `mb update --only-tools` | Só a fase **tools**: `mb tools --update-all`. A flag **`--only-tools`** só aparece na ajuda quando o agregador **`tools`** com **`--update-all`** está no cache (**`mb plugins sync`**). Em **`mb update`** sem flags, a fase tools ainda corre se o comando existir em runtime. **Só** `--only-tools` sem esse comando em runtime: fase ignorada com aviso. |
| `mb update --only-system` / `mb machine update` | Só a fase de pacotes do sistema: o MB executa **`mb machine update`** (scripts do plugin). Só disponível quando **`machine/update`** está instalado e sincronizado no cache; pode combinar com outros `--only-*`. Pode pedir password ao `sudo` no Linux (APT e Snap). |
| `mb plugins sync [--no-remove]` | Rescaneia o diretório de plugins e paths locais, atualiza o cache SQLite e garante os helpers de shell. Regista um **digest por comando** (manifest + ficheiros referenciados na folha); só emite linhas para comandos **novos**, **atualizados** (digest alterado) ou **removidos** do pacote. Se nada mudou nesses termos, uma linha curta indica cache atualizado sem alterações. Com **`--no-remove`**, mantém no cache entradas de comandos que já não existem na árvore (órfãos; `exec_path` pode ficar inválido). |
| `mb update --only-cli` | **Só para binários da release oficial** (versão embutida via ldflags no GitHub Release). Builds locais ou `go install` mostram mensagem a usar `install.sh`. Se a release for mais nova, baixa o `mb`, valida SHA256 e substitui o executável (Linux/macOS, amd64/arm64). |
| `mb update --only-cli --check-only` | Igual: só em binários de release. Compara com a última release (sem download). **Códigos de saída:** `0` = já atualizado ou versão local mais nova; `2` = há atualização; `1` = erro. Em build local: mensagem + saída `0`. Com **`--json`** (só com **`--only-cli --check-only`**), imprime no stdout uma linha JSON: `localVersion`, `remoteVersion`, `updateAvailable` (boolean alinhado à mesma regra que o código de saída). |
| `mb plugins add <url \| path \| .> [--package P] [--tag TAG] [--no-remove]` | Instala ou **substitui** um pacote com o mesmo identificador (re-clone ou atualiza `local_path`). **`--no-remove`** repassa ao sync (ver `mb plugins sync`). |
| `mb plugins list [--check-updates]` | Lista plugins instalados (pacote, comando, descrição, versão, **ORIGEM** (local/remoto), URL/path) |
| `mb plugins remove <package>` | Remove um pacote instalado (com confirmação). Se for local, só remove o registro. O cache é atualizado e o plugin deixa de aparecer em `plugins list`. |
| `mb plugins update [package \| --all]` | Atualiza um pacote remoto ou todos (plugins locais não são atualizados) |
| `mb envs groups [--json \| -J]` | Tabela GRUPO / ARQUIVO (`default` → `env.defaults`; mais uma linha por `.env.<grupo>` no config). **`--json` / `-J`**: array `[{"group","path"},...]`. Alias: **`group`**. |
| `mb envs list [--group G] [--show-secrets] [--json \| -J] [--text \| -T]` | Por omissão: tabela (VAR, GRUPO, ARMAZENAMENTO: `local`, `keyring`, `1password`); secrets mostram `***`; com `--show-secrets` mostra o valor (keyring; referências `op://` resolvidas com 1Password CLI). **`--json` / `-J`**: objeto JSON `{"CHAVE":"valor",...}`. **`--text` / `-T`**: uma linha `CHAVE=valor` por variável (sem coluna de grupo). **`--json` e `--text` são mutuamente exclusivos.** |
| `mb envs set <KEY> <VALUE> [--group G] [--secret \| --secret-op]` | Define em `env.defaults` ou em `.env.G`. **`--secret`**: valor no keyring. **`--secret-op`**: valor no 1Password, referência `op://` no keyring (exige `op` no PATH; mutuamente exclusivo com `--secret`). |
| `mb envs unset <KEY> [--group G]` | Remove de `env.defaults` ou de `.env.G` e, se for secret, do keyring (e 1Password quando usado). Com **`--group`**, se o grupo ficar sem vars nem `.secrets`, apaga **`.env.G`** (e **`.env.G.secrets`**). Se a chave não existir nesse grupo, mensagem informativa e saída **0** (nada alterado). |
| `mb run <comando> [args...]` | Executa um programa no PATH (ou caminho) com o **mesmo ambiente mesclado** que os plugins: `env.defaults`, `--env-group`, `./.env` no cwd (se existir), `--env-file`, `--env`, tema gum, `MB_HELPERS_PATH`, etc. **Sem** `env_files` de manifest (não há plugin). Stdin/stdout/stderr do terminal; **código de saída do filho propagado**. Flags globais (`-e`, `--env-file`, `--env-group`, `-v`/`-q`, etc.) podem ir **antes** de `run` ou **logo após** `run` (antes do executável); o resto vai para o filho. Ajuda: `mb help run` (evite `mb run --help`, que pode ir para o programa filho). Detalhes em [Variáveis de ambiente](../guide/environment-variables.md). |
| `mb <categoria> <comando> [args...]` | Executa o plugin correspondente (veja [Comandos de plugins](../guide/plugin-commands.md)) |

## Completion de shell

O CLI gera scripts de completion para bash, zsh, fish e powershell via `mb completion <shell>`. O completion inclui os comandos built-in e **todos os comandos e subcomandos de plugins** disponíveis no cache. Após `mb plugins sync` (ou após instalar um plugin), use TAB para sugerir categorias e comandos de plugins.

**Instalação automática:** `mb completion install` deteta o shell pela variável `SHELL` (no Windows assume PowerShell se `SHELL` não estiver definida), gera o mesmo script que `mb completion <shell>` e acrescenta (ou substitui) um bloco idempotente no ficheiro de perfil — por omissão `~/.bashrc`, `~/.zshrc`, `~/.config/fish/config.fish` (ou `XDG_CONFIG_HOME/fish/config.fish`), ou o perfil PowerShell do utilizador. Flags úteis: `--shell` para forçar o shell, `--rc-file` para outro ficheiro, `--dry-run` para pré-visualizar sem gravar, `-y` / `--yes` para gravar sem confirmação (necessário fora de um terminal interativo), `--no-descriptions` alinhado aos outros subcomandos de completion. Se o login shell usar apenas `.zprofile` em vez de `.zshrc`, use `--rc-file`. **Remover:** `mb completion uninstall` (mesmas ideias de flags) apaga só o bloco mb-cli do perfil; não falha se já não existir.

Depois de `mb completion install` no perfil **por omissão**, cada `mb plugins sync` (e operações que fazem sync, como `add` / `remove` / `update`) pode **atualizar o script embutido** automaticamente para refletir o cache atual — exceto se tiver instalado apenas com `--rc-file` para um caminho personalizado (nesse caso volte a correr `mb completion install`).

Para gerar só o script em stdout (instalação manual), consulte `mb completion --help` e os subcomandos `bash`, `zsh`, `fish`, `powershell`.

## Flags globais

- **`--verbose` / `-v`** — Saída mais verbosa. Veja [Flags globais](../guide/global-flags.md).
- **`--quiet` / `-q`** — Reduz mensagens. Veja [Flags globais](../guide/global-flags.md).
- **`--env-file <path>`** — Arquivo de variáveis de ambiente ao executar plugins ou **`mb run`**. Veja [Variáveis de ambiente](../guide/environment-variables.md).
- **`--env KEY=VALUE`** — Injeta variável no processo do plugin ou do **`mb run`** (pode ser repetido). Veja [Variáveis de ambiente](../guide/environment-variables.md).
- **`--env-group <nome>`** — Sobrepõe `env.defaults` com `~/.config/mb/.env.<nome>` ao executar plugins ou **`mb run`**. Veja [Variáveis de ambiente](../guide/environment-variables.md).
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
