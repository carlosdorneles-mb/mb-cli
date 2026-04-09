---
sidebar_position: 2
---

# Plugins

Esta página descreve como o MB CLI descobre, armazena e executa plugins — diretório de plugins, cache, sync e resolução de paths. Para **criar** um plugin e usar `mb plugins` no dia a dia, veja o [Guia: Criar um plugin](../guide/creating-plugins.md) e [Comandos de plugins](../guide/plugin-commands.md).

## Diretório de plugins e plugins locais

O MB usa um diretório de plugins derivado de `os.UserConfigDir()`:

- **Linux**: `~/.config/mb/plugins`
- **macOS**: `~/Library/Application Support/mb/plugins`

O **cache SQLite** do MB fica em **`ConfigDir/cache.db`** (o mesmo `ConfigDir` que `~/.config/mb` no Linux), **não** dentro da pasta de plugins.

**Plugins locais:** com `mb plugins add <path|.>` o path fica em `plugin_sources.local_path`; nada é copiado para `PluginsDir`. No sync, esse path é escaneado como uma fonte extra.

## Descoberta: scanner e `manifest.yaml`

O **scanner** percorre a árvore à procura de `manifest.yaml`. Para cada ficheiro:

**Validação (alinhada ao código):**

- Com **entrypoint**: o ficheiro tem de existir e estar **dentro** do diretório do plugin; o entrypoint não pode apontar para fora.
- Com **entrypoint** ou **flags**: o campo **`command`** no manifest é **obrigatório**.
- **Readme** e ficheiros em **env_files**: paths relativos ao plugin e validados sob o diretório do plugin.
- **flags** (entrypoints por flag): cada entrypoint declarado tem de estar sob o diretório do plugin.
- **flags.envs**: quando definido, deve ser lista de pares `KEY=VALUE` válidos.

Manifestos inválidos geram **avisos** no sync; o manifest é ignorado.

**Tipo no cache (`sh` / `bin`):** não há campo `type` obrigatório no manifest. O scanner define o tipo pelo entrypoint: termina em `.sh` → `sh`; caso contrário → `bin`.

### `command_path` e tipos de manifest

Sob a **raiz da fonte** (cada subpasta imediata de `PluginsDir`, ou o `local_path` registado), o CLI monta o `command_path`:

- Em cada nível de pasta usa `manifest.command` se existir; senão o nome da pasta.
- Na folha, o último segmento é o nome da pasta do plugin (ou `command` se for folha na raiz com entrypoint/flags).

Tipos:

- **Folha com entrypoint** — um executável por comando.
- **Folha flags-only** — só `flags`; o CLI escolhe o entrypoint pela flag.
- **Categoria** — sem entrypoint e sem `flags`; aparece como subcomando intermédio (descrição, opcionalmente README com `-r`). O campo **`aliases`** do manifest é gravado no cache (`categories.aliases_json`) e o Cobra regista os mesmos aliases que para uma folha (ex.: `mb pai <alias>` em vez do nome do segmento).

**Modos de scan:**

- **Scan(`PluginsDir`)** — Para cada subdiretório de `PluginsDir`, essa pasta é a raiz da árvore. O `command_path` **não** inclui o nome da pasta clone.
- **ScanDir(`rootPath`, `installName`)** — O segundo argumento **não** entra no path do CLI; serve para identificar a fonte em `plugin_sources`. Usado para cada `local_path`.

São guardados `plugin_dir`, `exec_path` e `readme_path` **absolutos** quando aplicável.

## Grupos de help (`groups.yaml`, `group_id` e Cobra)

### Ficheiro `groups.yaml`

Pode existir em **qualquer pasta** do pacote. Formato: lista de pares `id` / `title`:

```yaml
- id: meu_grupo
  title: TÍTULO NO HELP
```

Regras de **`id`** (validação ao carregar o ficheiro):

- Deve coincidir com `^[a-zA-Z][a-zA-Z0-9_]*$` — **começa com letra**, depois letras, dígitos ou `_`.
- Não pode ser `commands` nem `plugin_commands` (reservados para o Cobra).
- **`title`** não pode ser vazio.
- O mesmo **`id` repetido dentro do mesmo ficheiro** é **erro** (o ficheiro é rejeitado com aviso no sync).

### Merge no sync e registo global

No **sync**, cada `groups.yaml` válido é um **lote**. Dentro de cada árvore, os paths dos ficheiros são processados em **ordem lexicográfica**. A ordem global dos lotes é:

1. Cada subpasta de `PluginsDir`, na ordem devolvida pelo sistema de ficheiros (`ReadDir`).
2. Depois, cada fonte com `local_path` não vazio, na ordem de `ListPluginSources()`.

Para o mesmo **`id`** em lotes diferentes: prevalece a **última** definição. Se o **título** mudar relativamente ao valor já fixado, o CLI regista **log debug** (visível com **`mb -v`**, via gum quando disponível).

O resultado é gravado na tabela **`plugin_help_groups`** (`group_id` → `title`). Qualquer folha ou categoria **aninhada** pode referenciar **qualquer** `id` presente nesse registo após o merge (não está limitado ao `groups.yaml` do pai imediato).

### Campo `group_id` no `manifest.yaml`

Campo opcional YAML: `group_id`.

- Só é usado para entradas **aninhadas**, isto é, quando o path tem **`/`** (`command_path` de plugin ou `path` de categoria, ex. `infra/k8s`).
- **Top-level** (sem `/`): `group_id` é **ignorado** e gera log **debug** se estiver preenchido.
- O manifest **não** valida o formato de `group_id` como o `groups.yaml`; convém usar o mesmo padrão de `id`. No **sync**, se o `group_id` **não** existir em `plugin_help_groups` após o merge, o campo é **removido** no cache e é emitido log **debug**. O sync **não falha**.

### Como o help aparece no CLI

Ao registar comandos a partir do cache:

- Comandos cuja categoria está **diretamente sob a raiz `mb`** ficam no grupo **COMANDOS DE PLUGINS** (id interno `plugin_commands`).
- Subcomandos **aninhados** (ex. `mb infra ci`):
  - Com `group_id` válido no cache → secção com o **título** definido em `plugin_help_groups`.
  - Sem `group_id` ou inválido no cache → secção **COMANDOS** (id `commands`).
- Nos comandos **pai** aninhados, o MB também regista no Cobra os grupos custom do registo, para as folhas herdarem as mesmas secções.

## Cache SQLite e sync

### Tabelas relevantes

- **plugins** — Entre outros: `command_path`, `command_name`, `plugin_dir`, descrição, `exec_path`, tipo (`sh`/`bin` ou vazio), `config_hash`, `readme_path`, `flags_json`, `env_files_json`, `hidden`, **`group_id`** (só faz sentido para paths aninhados; valores inválidos são limpos no sync).
- **categories** — `path`, descrição, `readme_path`, `hidden`, **`group_id`** (categorias aninhadas; mesma regra de validação que nos plugins).
- **plugin_help_groups** — `group_id` → `title` (resultado do merge dos `groups.yaml`).
- **plugin_sources** — Por `install_dir`: remoto (`git_url`, ref, versão) ou **`local_path`** para plugins adicionados localmente.

### Fluxo do sync (`mb plugins sync` e após add/remove/update de plugins)

Ordem **efetiva** no código:

1. **Scan** de todas as subpastas de `PluginsDir` → plugins, categorias, avisos e lotes de `groups.yaml`.
2. Garantir **helpers de shell** em `ConfigDir/lib/shell` (ficheiros `*.sh` embebidos no binário e `.checksum`; em cada sync reescreve-se e removem-se `*.sh` órfãos). Falha → erro.
3. **ScanDir** de cada `local_path` registado → acrescenta plugins, categorias, avisos e lotes.
4. **Merge global** dos grupos de help; em seguida, emissão de **avisos** de scan (warn) se houver logger.
5. Verificação de **colisão de `command_path`** entre pacotes distintos → **erro** e sync abortado.
6. **Normalização de `group_id`**: plugins e categorias aninhados com `group_id` inexistente no merge são limpos (+ debug).
7. **Apagar** todos os plugins e todas as linhas de `plugin_help_groups`; **inserir** o merge em `plugin_help_groups`; **upsert** de cada plugin.
8. **Apagar** todas as categorias; **upsert** de cada categoria.

**Nota:** `plugin_sources` **não** é alterado pelo sync; só por `mb plugins add/remove/update`.

A árvore de comandos no CLI reflete `PluginsDir` **mais** os paths locais registados.

### `config_hash` (digest da folha)

Para cada manifest de **folha** (com `entrypoint` ou só `flags`), o `config_hash` é um SHA-256 hexadecimal de um bloco canónico: inclui o hash do `manifest.yaml` bruto e, por ficheiro referenciado e ordenado por path relativo, `path<TAB>SHA256(conteúdo)`. Ficheiros considerados: `entrypoint`, cada `entrypoint` em `flags`, cada `file` em `env_files`, e `readme` se definido. Não há walk recursivo do diretório.

No **diff** do sync (com logger), só são emitidas linhas para comandos **adicionados**, com digest **alterado** em relação ao cache, ou **removidos** da árvore; comandos inalterados não geram linha. Com **`EmitSuccess`** (ex.: `mb plugins sync`), se não houver nenhuma alteração nesses termos, é mostrada uma única mensagem de resumo em vez do antigo contador genérico de plugins. A lógica de contagem é a mesma usada pelo CLI em **`mb plugins add`** para decidir se mostra a mensagem genérica de “pacote já alinhado” (nenhum comando novo, atualizado ou removido).

## Execução: binário ou script

- **Com entrypoint:** o cache guarda `ExecPath` absoluto. Se o ficheiro termina em `.sh`, o executor usa **bash** com esse script; senão executa o ficheiro diretamente.
- **Flags-only:** o diretório de trabalho é o `plugin_dir`; o entrypoint de cada flag é resolvido dentro dele.

**Indicação (local)** no Short da folha quando a fonte tem `local_path` em `plugin_sources` (match pelo diretório do plugin).

### Códigos de saída convencionais para scripts shell {#plugin-shell-exit-codes-convention}

Plugins que participam em batches como **`mb tools --update-all`** (ou equivalentes) podem usar códigos acordados com o script agregador:

- **86** (`MB_EXIT_UPDATE_SKIPPED_SUDO`) — sem `sudo` não interativo / root para operações de pacote; o batch trata como “saltado”, não como erro fatal.
- **87** (`MB_EXIT_UPDATE_SKIPPED_NOT_INSTALLED`) — nada a atualizar porque a ferramenta não está instalada; ignorado no batch.

O executor Go repassa o código de saída do subprocesso; mensagens amigáveis para **86** em invocação direta ficam a cargo dos scripts (helper **`warn_and_skip_without_sudo`** em `~/.config/mb/lib/shell/sudo.sh`, sincronizado com `mb plugins sync`). Ver [Guia: criar um plugin — códigos e sudo](../guide/creating-plugins.md#plugin-exit-codes-sudo).

## Execução: flags e argumentos

O processo do plugin **não recebe as flags tratadas pelo CLI**; recebe apenas **argumentos posicionais** (e o ambiente mesclado, incl. `MB_VERBOSE` / `MB_QUIET` quando aplicável).

### Flags que o CLI consome

| Origem | Flags | Efeito |
|--------|--------|--------|
| **Root** | `-v` / `--verbose`, `-q` / `--quiet`, `--env-file`, `-e` / `--env` | Consumidas pelo CLI; `-v`/`-q` viram variáveis de ambiente do plugin. |
| **Plugin com README** | `-r` / `--readme` | Abre a documentação no terminal. |
| **Plugin com `flags` no manifest** | Flags declaradas | Escolhem o entrypoint; não são repassadas como args ao script; `flags[].envs` só é injetado para flags efetivamente informadas. |

### O que o script recebe

- **Posicionais** após o consumo das flags acima. Ex.: `mb tools hello foo bar` → `$1`, `$2`.
- **Ambiente:** sistema + defaults + `env_files` do plugin + `flags[].envs` (apenas das flags usadas) + `--env` + `MB_VERBOSE`/`MB_QUIET` + variáveis de contexto **`MB_CTX_*`** (invocação, path no manifest, irmãos Cobra, etc.).
  - Precedência prática em conflito de chave: `--env` > `flags[].envs` > `env_files`.
  - Ver [Variáveis de ambiente](../guide/environment-variables.md), [Flags globais](../guide/global-flags.md) e [Contexto de invocação de plugins](plugin-invocation-context.md).

### `--help` / `-h`

Mostra o help Cobra do comando; o entrypoint **não** é executado.

### Flags desconhecidas

- **Um único entrypoint:** globais e opcionalmente `--readme` são consumidas pelo CLI; o restante depende do parser — para compatibilidade, prefira posicionais ou declare flags no manifest.
- **Plugin com `flags` no manifest:** flags não declaradas provocam erro *unknown flag* e o plugin não corre.

## Segurança

Os plugins correm com as permissões do utilizador; o CLI confina paths ao diretório do plugin no scan e no executor e suporta timeout opcional. Ver [Segurança](../guide/security.md).
