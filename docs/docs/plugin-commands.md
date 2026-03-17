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

## Como descobrir os comandos

- **`mb plugins list`** — Lista todos os plugins instalados, com nome, caminho do comando, descrição, versão, origem (local ou remoto) e URL/path. Use essa saída para saber exatamente quais comandos estão disponíveis.
- **`mb help`** — Mostra a árvore de comandos, incluindo as categorias e comandos de plugins. Comandos de plugins locais aparecem com a indicação "(local)" na descrição.
- **Completion** — Depois de rodar `mb self sync`, o completion do shell (TAB) sugere categorias e comandos de plugins. Instale com `mb self completion <bash|zsh|fish|powershell>`.

## Executando um comando de plugin

Basta chamar o comando com os argumentos que o plugin espera:

```bash
mb tools hello
mb infra ci deploy --ambiente prod
```

Se o plugin declarou um README no manifesto, o comando ganha a flag **`--readme`** (ou `-r`), que abre a documentação em Markdown no terminal (via glow, se instalado):

```bash
mb tools meu-comando --readme
```

Para detalhes do que acontece com **flags e argumentos** ao chamar um comando de plugin (quais flags o CLI consome, o que o script recebe em `$1`, `$2`, etc., e o que ocorre quando se passam flags que não existem), veja a seção [Execução: flags e argumentos passados ao plugin](./plugins.md#execução-flags-e-argumentos-passados-ao-plugin) na referência técnica.

## Repositório com vários plugins

Um único `mb plugins add <url>` ou `mb plugins add <path>` cobre **toda a árvore** do diretório. Os comandos no CLI seguem a hierarquia de pastas e os `manifest.yaml` (campo `command` por nível quando quiser renomear um segmento), **sem** prefixar pelo nome da instalação. Exemplo: repo com `tools/postman` e `dev/kinfo` → `mb tools postman`, `mb dev kinfo`. Em **`mb plugins list`**, a coluna **NOME** é o identificador da instalação (`--name` ou nome do diretório clone), usado em `mb plugins remove <nome>`.

## Plugin local vs remoto

Na listagem (`mb plugins list`), a coluna **ORIGEM** indica se o plugin é **local** (instalado por path ou `.`) ou **remoto** (instalado por URL Git). No help (`mb help` ou `mb <categoria> <comando> --help`), comandos de plugins locais exibem **(local)** ao lado da descrição, para você saber que aquele comando vem de um plugin registrado localmente.

Para detalhes de como o CLI descobre e executa os plugins (cache, sync, resolução de paths), veja [Plugins (referência técnica)](./plugins.md).
