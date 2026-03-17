---
sidebar_position: 2
---

# Plugins

Esta página descreve como o MB CLI descobre, armazena e executa plugins — diretório de plugins, cache, sync e resolução de paths. Para como **criar** um plugin e usar os comandos `mb plugins` no dia a dia, veja o [Guia: Criar um plugin](../guide/creating-plugins.md) e [Comandos de plugins](../guide/plugin-commands.md).

## Diretório de plugins e plugins locais

O MB usa um único diretório de plugins derivado de `os.UserConfigDir()`:

- **Linux**: `~/.config/mb/plugins`
- **macOS**: `~/Library/Application Support/mb/plugins`

Além desse diretório, o CLI suporta **plugins locais**: em vez de clonar ou copiar para `PluginsDir`, o usuário pode registrar um path do sistema de arquivos (ou `.` para o diretório atual) com `mb plugins add <path|.>`. Nesse caso, nada é copiado para `PluginsDir`; o path fica gravado em `plugin_sources.local_path` e o conteúdo é escaneado a partir desse path no sync.

## Descoberta: scanner e manifest.yaml

O **scanner** percorre um diretório em busca de arquivos `manifest.yaml`. Para cada manifesto encontrado:

- Validação: `type` deve ser `sh` ou `bin` se houver `entrypoint`; o arquivo do entrypoint deve existir.
- A partir da árvore sob a **raiz da fonte** (cada subdiretório imediato de `PluginsDir`, ou o `local_path` registrado), o CLI monta o `command_path`: em cada nível usa `manifest.command` se existir, senão o nome da pasta; na folha, o último segmento é o nome da pasta do plugin.
- Cada manifesto pode definir um **plugin** (com entrypoint ou com `flags`) ou apenas uma **categoria** (sem entrypoint e sem flags).

Dois modos de scan:

- **Scan(pluginsDir)** — Para cada subdiretório de `PluginsDir` (cada clone/cópia), percorre essa árvore como raiz. O `command_path` **não** inclui o nome da pasta do clone.
- **ScanDir(rootPath, installName)** — O segundo parâmetro é ignorado para o path no CLI; percorre `rootPath` (ex.: `local_path`) com a mesma regra. `installName` identifica só a fonte em `plugin_sources`.

Os resultados guardam `plugin_dir` (pasta do manifest), `ExecPath` e `ReadmePath` absolutos.

## Cache e sync

O cache SQLite (`cache.db`) armazena:

- **plugins** — command_path, command_name, plugin_dir, descrição, exec_path, tipo, config_hash, readme_path, flags_json; e campos Cobra (use_template, args_count, etc.).
- **categories** — Path, descrição, readme_path.
- **plugin_sources** — Por install_dir: git_url, ref, version (para remotos) ou local_path (para locais).

O **sync** (`mb self sync` ou após add/remove/update):

1. Garante que os **helpers de shell** existem em `ConfigDir/lib/shell` (cria ou atualiza `all.sh`, `log.sh` e `.checksum`). Se falhar (ex.: permissão), o sync retorna erro.
2. Chama o scanner em `PluginsDir` e obtém listas de plugins e categories.
3. Obtém `ListPluginSources()`; para cada source com `local_path` não vazio, chama `ScanDir` nesse path e faz **merge**. Se dois pacotes definem o mesmo `command_path`, o sync **falha** com erro de conflito.
4. Faz upsert de todos os plugins e categories no cache.
5. **plugin_sources** só é alterado por `plugins add/remove/update`, não pelo sync a partir dos comandos descobertos.

Assim, a árvore de comandos reflete tanto o conteúdo de `PluginsDir` quanto dos diretórios registrados como locais.

## Execução: como o binário/script é localizado

- Para plugins com **entrypoint** (um único script ou binário): o cache já guarda `ExecPath` absoluto (preenchido pelo scanner). O executor recebe esse path e o ambiente mesclado e invoca o processo (quando o entrypoint termina em `.sh`, invoca **bash** com o script como argumento; caso contrário, executa o binário diretamente).
- Para plugins **flags-only**: o `baseDir` da execução é o **`plugin_dir`** guardado no cache (pasta do manifest). O `exec_path` da flag é resolvido dentro desse diretório.

A indicação **(local)** no Short do comando folha quando a fonte do plugin tem `local_path` em `plugin_sources` (match pelo diretório do plugin).

## Execução: flags e argumentos passados ao plugin

O processo do plugin (script ou binário) **nunca recebe flags na linha de comando**. O CLI trata as flags que conhece e repassa ao entrypoint apenas **argumentos posicionais** (como `$1`, `$2`, … no shell). O ambiente do processo inclui variáveis injetadas (por exemplo `MB_VERBOSE`, `MB_QUIET` quando se usa `-v` ou `-q`).

### Flags que o CLI conhece

| Origem | Flags | O que acontece |
|--------|--------|----------------|
| **Root (globais)** | `-v` / `--verbose`, `-q` / `--quiet`, `--env-file`, `-e` / `--env` | Consumidas pelo CLI. `-v` e `-q` não vão para o script; em troca, o CLI define no ambiente do plugin `MB_VERBOSE=1` ou `MB_QUIET=1`. Podem ser usadas em qualquer posição (ex.: `mb tools hello -v`). |
| **Plugin com README** | `-r` / `--readme` | Consumida pelo CLI; abre a documentação no terminal. Não é repassada ao script. |
| **Plugin com `flags` no manifesto** | As flags declaradas no manifesto (ex.: `--deploy`, `--rollback`) | Consumidas pelo CLI para **escolher qual entrypoint** rodar. Não são repassadas como argumentos ao script; apenas os argumentos posicionais restantes são passados. |

### O que o script/binário recebe

- **Argumentos:** apenas os **posicionais** que sobraram depois de o CLI consumir as flags acima. Exemplo: `mb tools hello foo bar` → o script recebe `foo` e `bar` em `$1` e `$2`. Se o usuário passar `mb tools hello -v`, o `-v` é consumido pelo CLI (e vira `MB_VERBOSE=1` no env); o script recebe **nenhum** argumento posicional.
- **Ambiente:** ambiente do sistema + arquivo de defaults + `--env-file` + `--env` + `MB_VERBOSE=1` e/ou `MB_QUIET=1` quando as flags globais forem usadas. Ver [Variáveis de ambiente](../guide/environment-variables.md) e [Flags globais](../guide/global-flags.md).

### Quando você passa flags que **existem** (conhecidas pelo CLI)

As flags listadas na tabela acima são reconhecidas e **não** aparecem em `$1`, `$2`, … O comportamento é o descrito: globais afetam o ambiente; `--readme` abre o README; flags do manifesto (no caso de plugin com `flags`) escolhem o entrypoint. Os argumentos posicionais restantes são os únicos passados ao entrypoint.

### --help / -h

Em **qualquer** comando de plugin, `--help` ou `-h` exibe o help do Cobra (descrição, uso, flags conhecidas) e **não** é repassado ao script. O entrypoint não é executado quando o usuário pede help.

### Quando você passa flags que **não existem**

- **Plugin com um único entrypoint (com ou sem README):** o CLI faz parsing das flags globais (`-v`, `-q`, `--env-file`, `-e`) e, se houver README, da flag `--readme`. Essas flags são sempre consumidas pelo CLI e não chegam ao script. Os **argumentos posicionais** restantes são repassados ao entrypoint. Flags não declaradas (desconhecidas) podem não ser repassadas ao script, dependendo do analisador de argumentos; para máxima compatibilidade, use apenas argumentos posicionais ou declare flags no manifesto.
- **Plugin com `flags` no manifesto:** o comando do plugin faz parsing das flags declaradas. Se o usuário passar uma flag que **não** está declarada (nem no root, nem no plugin), o Cobra retorna erro do tipo *unknown flag* e o plugin **não** é executado.

Resumindo: as flags globais e `--help` são sempre tratadas pelo CLI; apenas os argumentos posicionais (e, quando aplicável, flags não mapeadas) são repassados ao entrypoint.

## Segurança

Os plugins rodam com as permissões do usuário; o CLI restringe a execução a scripts **dentro do diretório do plugin** (confinamento de path no scan e no executor) e suporta timeout opcional. Para o modelo completo e recomendações, veja [Segurança](../guide/security.md).
