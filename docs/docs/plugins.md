---
sidebar_position: 2
---

# Plugins

Esta página descreve como o MB CLI descobre, armazena e executa plugins — diretório de plugins, cache, sync e resolução de paths. Para como **criar** um plugin e usar os comandos `mb plugins` no dia a dia, veja o [Guia: Criar um plugin](./creating-plugins.md) e [Comandos de plugins](./plugin-commands.md).

## Diretório de plugins e plugins locais

O MB usa um único diretório de plugins derivado de `os.UserConfigDir()`:

- **Linux**: `~/.config/mb/plugins`
- **macOS**: `~/Library/Application Support/mb/plugins`

Além desse diretório, o CLI suporta **plugins locais**: em vez de clonar ou copiar para `PluginsDir`, o usuário pode registrar um path do sistema de arquivos (ou `.` para o diretório atual) com `mb plugins add <path|.>`. Nesse caso, nada é copiado para `PluginsDir`; o path fica gravado em `plugin_sources.local_path` e o conteúdo é escaneado a partir desse path no sync.

## Descoberta: scanner e manifest.yaml

O **scanner** percorre um diretório em busca de arquivos `manifest.yaml`. Para cada manifesto encontrado:

- Validação: `type` deve ser `sh` ou `bin` se houver `entrypoint`; o arquivo do entrypoint deve existir.
- A partir do path relativo ao root escaneado, o CLI monta o `command_path` (ex.: `tools/hello`, `infra/ci/deploy`).
- Cada manifesto pode definir um **plugin** (com entrypoint ou com `flags`) ou apenas uma **categoria** (sem entrypoint e sem flags).

Dois modos de scan:

- **Scan(pluginsDir)** — Percorre apenas `PluginsDir`; usado para plugins instalados por Git (ou copiados manualmente) nesse diretório.
- **ScanDir(rootPath, installName)** — Percorre um path arbitrário (ex.: um `local_path`) e prefixa todos os `command_path` com `installName`. Usado no sync para cada plugin source que tem `local_path` preenchido.

Os resultados (plugins e categories) usam paths absolutos para `ExecPath` e `ReadmePath` quando vêm do ScanDir, de modo que a execução funcione mesmo quando o plugin está fora de `PluginsDir`.

## Cache e sync

O cache SQLite (`cache.db`) armazena:

- **plugins** — Comando, descrição, exec_path, tipo, config_hash, readme_path, flags_json; e, quando definidos no manifest, use_template, args_count, aliases_json, example, long_description, deprecated (para Cobra).
- **categories** — Path, descrição, readme_path.
- **plugin_sources** — Por install_dir: git_url, ref, version (para remotos) ou local_path (para locais).

O **sync** (`mb self sync` ou após add/remove/update):

1. Garante que os **helpers de shell** existem em `ConfigDir/lib/shell` (cria ou atualiza `all.sh`, `log.sh` e `.checksum`). Se falhar (ex.: permissão), o sync retorna erro.
2. Chama o scanner em `PluginsDir` e obtém listas de plugins e categories.
3. Obtém `ListPluginSources()`; para cada source com `local_path` não vazio, chama `ScanDir(local_path, installDir)` e faz **merge** (append) dos resultados.
4. Faz upsert de todos os plugins e categories no cache (replace por `command_path` ou `path`).
5. Atualiza **plugin_sources**: para cada top-level dir que apareceu no scan e ainda não tem linha no banco, insere uma linha (com git_url e local_path vazios). Linhas existentes são preservadas (incluindo `local_path` e `git_url`).

Assim, a árvore de comandos reflete tanto o conteúdo de `PluginsDir` quanto dos diretórios registrados como locais.

## Execução: como o binário/script é localizado

- Para plugins com **entrypoint** (um único script ou binário): o cache já guarda `ExecPath` absoluto (preenchido pelo scanner). O executor recebe esse path e o ambiente mesclado e invoca o processo (quando o entrypoint termina em `.sh`, invoca **bash** com o script como argumento; caso contrário, executa o binário diretamente).
- Para plugins **flags-only** (várias ações por flag): o cache guarda `flags_json`. O handler do comando folha sabe qual flag foi escolhida e qual entrypoint corresponde; o **plugin root** é obtido assim: se há `plugin_sources[installDir].LocalPath`, usa esse path; senão, usa `filepath.Join(PluginsDir, installDir)`. O `baseDir` do comando é o plugin root + o sufixo do `command_path` (segmentos após o primeiro). O `exec_path` efetivo é `baseDir + entrypoint` da flag.

A indicação **(local)** no Short do comando folha vem do fato de o plugin ter `local_path` preenchido em `plugin_sources`; o mesmo dado é usado para resolver o plugin root na execução.

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
- **Ambiente:** ambiente do sistema + arquivo de defaults + `--env-file` + `--env` + `MB_VERBOSE=1` e/ou `MB_QUIET=1` quando as flags globais forem usadas. Ver [Variáveis de ambiente](./environment-variables.md) e [Flags globais](./global-flags.md).

### Quando você passa flags que **existem** (conhecidas pelo CLI)

As flags listadas na tabela acima são reconhecidas e **não** aparecem em `$1`, `$2`, … O comportamento é o descrito: globais afetam o ambiente; `--readme` abre o README; flags do manifesto (no caso de plugin com `flags`) escolhem o entrypoint. Os argumentos posicionais restantes são os únicos passados ao entrypoint.

### Quando você passa flags que **não existem**

- **Plugin com um único entrypoint e sem README no manifesto:** o comando do plugin tem `DisableFlagParsing = true`. Nada é interpretado como flag; **tudo** que vier depois do nome do comando é repassado ao script como argumentos posicionais. Exemplo: `mb tools hello --foo=bar -x` → o script recebe `$1=--foo=bar`, `$2=-x`. O script pode interpretar isso como quiser (por exemplo, com `getopts` ou um parser próprio).
- **Plugin com README ou com `flags` no manifesto:** o comando do plugin faz parsing de flags (só as que o CLI declarou). Se o usuário passar uma flag que **não** está declarada (nem no root, nem no plugin), o Cobra retorna erro do tipo *unknown flag* e o plugin **não** é executado.

Resumindo: em plugins “simples” (um entrypoint, sem README), qualquer coisa após o comando vira argumento do script. Em plugins com README ou com flags no manifesto, apenas as flags conhecidas são aceitas; o resto gera erro antes de rodar o plugin.

## Segurança

Os plugins rodam com as permissões do usuário; o CLI restringe a execução a scripts **dentro do diretório do plugin** (confinamento de path no scan e no executor) e suporta timeout opcional. Para o modelo completo e recomendações, veja [Segurança](./security.md).
