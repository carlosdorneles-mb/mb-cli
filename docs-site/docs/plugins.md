---
sidebar_position: 2
---

# Plugins (referência técnica)

Esta página descreve como o MB CLI descobre, armazena e executa plugins — diretório de plugins, cache, sync e resolução de paths. Para como **criar** um plugin e usar os comandos `mb plugins` no dia a dia, veja o [Guia: Criar um plugin](./creating-plugins.md) e [Comandos de plugins](./comandos-plugins.md).

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

- **plugins** — Comando, descrição, exec_path, tipo, config_hash, readme_path, flags_json.
- **categories** — Path, descrição, readme_path.
- **plugin_sources** — Por install_dir: git_url, ref, version (para remotos) ou local_path (para locais).

O **sync** (`mb self sync` ou após add/remove/update):

1. Chama o scanner em `PluginsDir` e obtém listas de plugins e categories.
2. Obtém `ListPluginSources()`; para cada source com `local_path` não vazio, chama `ScanDir(local_path, installDir)` e faz **merge** (append) dos resultados.
3. Faz upsert de todos os plugins e categories no cache (replace por `command_path` ou `path`).
4. Atualiza **plugin_sources**: para cada top-level dir que apareceu no scan e ainda não tem linha no banco, insere uma linha (com git_url e local_path vazios). Linhas existentes são preservadas (incluindo `local_path` e `git_url`).

Assim, a árvore de comandos reflete tanto o conteúdo de `PluginsDir` quanto dos diretórios registrados como locais.

## Execução: como o binário/script é localizado

- Para plugins com **entrypoint** (um único script ou binário): o cache já guarda `ExecPath` absoluto (preenchido pelo scanner). O executor recebe esse path e o ambiente mesclado e invoca o processo (para `type: sh`, typically `/bin/sh` + script como argumento).
- Para plugins **flags-only** (várias ações por flag): o cache guarda `flags_json`. O handler do comando folha sabe qual flag foi escolhida e qual entrypoint corresponde; o **plugin root** é obtido assim: se há `plugin_sources[installDir].LocalPath`, usa esse path; senão, usa `filepath.Join(PluginsDir, installDir)`. O `baseDir` do comando é o plugin root + o sufixo do `command_path` (segmentos após o primeiro). O `exec_path` efetivo é `baseDir + entrypoint` da flag.

A indicação **(local)** no Short do comando folha vem do fato de o plugin ter `local_path` preenchido em `plugin_sources`; o mesmo dado é usado para resolver o plugin root na execução.
