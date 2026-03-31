---
sidebar_position: 3
---

# Comandos de plugins

Os comandos de plugins são aqueles que o MB CLI monta automaticamente a partir dos plugins instalados. Eles aparecem como `mb <categoria> <comando>` (e podem ter subcategorias, por exemplo `mb infra ci deploy`).

## O que são

Cada plugin que você instala ou registra vira um ou mais comandos na árvore do CLI. A **categoria** vem da estrutura de pastas (ou do path do plugin); o **comando** é o nome definido no `manifest.yaml` ou o nome da pasta. Assim, você executa o plugin chamando o comando correspondente.

Exemplos:

- Plugin em `tools/hello/` → `mb tools hello`
- Plugin em `infra/ci/deploy/` → `mb infra ci deploy`

Na folha com **entrypoint** ou **flags**, o nome do subcomando no CLI vem do campo **`command`** do `manifest.yaml` (o último segmento do caminho interno continua a ser o nome da pasta). Plugins **só com `flags`** (sem entrypoint raiz) precisam de uma **flag declarada** para correr um script — ex.: `mb tools do --deploy`.

**Cache:** os comandos vêm do SQLite após **`mb plugins sync`**. O `mb plugins add` dispara o sync; se editar ficheiros diretamente em `PluginsDir`, volte a correr **`mb plugins sync`** para atualizar listagem, help e completion. O MB compara um **digest por comando** (folha): só aparecem mensagens por comando para **adicionados**, **alterados** (conteúdo do plugin / digest mudou) ou **removidos** do pacote; **não** há linha de log quando o comando já existia e o digest não mudou. Em **`mb plugins sync`**, se não houver nenhuma dessas alterações, verá uma linha curta a indicar que o cache foi atualizado sem mudanças nos comandos. Em **`mb plugins add`**, se o pacote já estiver alinhado (nenhum comando novo, atualizado ou removido), é mostrada uma **mensagem genérica** em vez de repetir confirmações por comando. Use **`--no-remove`** no sync/add para **manter** no cache comandos já ausentes da árvore (com aviso).

## Como descobrir os comandos

- **`mb plugins list`** — Lista todos os plugins instalados, com pacote (identificador da instalação), caminho do comando, descrição, versão, origem (local ou remoto) e URL/path. Use essa saída para saber exatamente quais comandos estão disponíveis.
- **`mb help`** — Mostra a árvore de comandos, incluindo as categorias e comandos de plugins. Comandos de plugins locais aparecem com a indicação "(local)" na descrição.
- **Completion** — Depois de `mb plugins sync`, o completion (TAB) sugere categorias e comandos. Instale no perfil com `mb completion install`, remova com `mb completion uninstall`, ou gere o script com `mb completion <bash|zsh|fish|powershell>` (ver `mb completion install --help`).

No help (`mb help`), subcomandos aninhados podem aparecer em **COMANDOS** ou em secções definidas com `groups.yaml` / `group_id` — ver [Grupos de help](../technical-reference/plugins.md#grupos-de-help-groupsyaml-group_id-e-cobra).

## Executando um comando de plugin

Basta chamar o comando com os argumentos que o plugin espera:

```bash
mb tools hello
mb infra ci deploy --ambiente prod
```

Com **`readme`** no manifest, a **folha** e também uma **categoria** (manifest sem entrypoint/flags) podem expor **`--readme`** / **`-r`** para ver o Markdown no terminal (glow, se instalado):

```bash
mb tools meu-comando --readme
mb infra --readme
```

Para **flags globais do CLI**, argumentos posicionais no script e flags desconhecidas, veja [Execução: flags e argumentos](../technical-reference/plugins.md#execução-flags-e-argumentos).

## Repositório com vários plugins

Um único `mb plugins add <url>` ou `mb plugins add <path>` cobre **toda a árvore** do diretório. Os comandos no CLI seguem a hierarquia de pastas e os `manifest.yaml` (campo `command` por nível quando quiser renomear um segmento), **sem** prefixar pelo identificador do pacote na árvore de comandos. Exemplo: repo com `tools/postman` e `dev/kinfo` → `mb tools postman`, `mb dev kinfo`. Em **`mb plugins list`**, a coluna **PACOTE** é o identificador da instalação (`--package` ou nome do diretório clone), usado em `mb plugins remove <package>` e `mb plugins update <package>`.

## Plugin local vs remoto

Na listagem (`mb plugins list`), a coluna **ORIGEM** indica se o plugin é **local** (instalado por path ou `.`) ou **remoto** (instalado por URL Git). No help (`mb help` ou `mb <categoria> <comando> --help`), comandos de plugins locais exibem **(local)** ao lado da descrição, para você saber que aquele comando vem de um plugin registrado localmente.

Para detalhes de como o CLI descobre e executa os plugins (cache, sync, resolução de paths), veja [Plugins (referência técnica)](../technical-reference/plugins.md).
