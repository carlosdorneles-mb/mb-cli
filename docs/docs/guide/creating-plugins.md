---
sidebar_position: 2
---

# Criar um plugin

Guia para montar um pacote de plugins do MB CLI: pastas, `manifest.yaml`, registo e execução.

- **Referência técnica** (scanner, cache, sync, grupos de help): [Plugins](../technical-reference/plugins.md)
- **Uso no dia a dia** (`mb plugins`, help, completion): [Comandos de plugins](./plugin-commands.md)
- **Exemplos no repositório**: [examples/plugins](https://github.com/carlosdorneles-mb/mb-cli/tree/main/examples/plugins) — na raiz do repo, `make install-examples` e depois `mb plugins sync`

## Checklist rápido

1. Árvore de pastas com `manifest.yaml` em cada nível necessário (categorias + folhas).
2. Folha executável: `command` + `entrypoint` (ou só `flags` para modo flags-only).
3. Registar: `mb plugins add …` (local ou Git) **ou** copiar para `~/.config/mb/plugins/<nome>/`.
4. **`mb plugins sync`** (automático após `plugins add`; obrigatório se alterou ficheiros à mão).
5. Testar: `mb plugins list` e o comando na CLI.

## Onde colocar o pacote

| Forma | O que acontece |
|-------|----------------|
| **Local** — `mb plugins add <path>` ou `mb plugins add .` | O path fica em `plugin_sources.local_path`. **Nada é copiado** para a pasta de plugins. O sync lê esse diretório. |
| **Remoto** — `mb plugins add <url-git>` | Clone para `PluginsDir/<nome>` (`--name` define o nome da instalação; `--tag` fixa uma tag). |
| **Manual** — criar/copiar pastas em `~/.config/mb/plugins/<nome>/` (Linux) | Depois **`mb plugins sync`** para atualizar o cache. |

## Árvore de pastas e caminho no CLI

A **raiz da fonte** é: cada subpasta imediata de `PluginsDir`, **ou** o diretório registado como `local_path`.

- Em cada nível, o segmento do comando vem de **`command`** no `manifest.yaml` da pasta, se existir; senão do **nome da pasta**.
- Na **folha** (plugin executável), o último segmento do path interno é o **nome da pasta**; o nome do subcomando Cobra vem do **`command`** do manifest (obrigatório com `entrypoint` ou `flags`).

Exemplos:

- `tools/hello/manifest.yaml` com `command: hello` → **`mb tools hello`**
- `infra/ci/deploy/` (folha) → **`mb infra ci deploy`**

Se **duas fontes** distintas expuserem o mesmo **`command_path`**, o **`mb plugins sync`** **falha** até remover ou ajustar uma delas.

## Tipos de `manifest.yaml`

### Categoria (sem folha executável)

Sem `entrypoint` e sem lista `flags`. Define um subcomando intermédio (descrição, opcionalmente `readme` e flag `-r`).

Campos úteis: `command`, `description`, `long_description`, `readme`, `hidden`.  
Para **help agrupado** em comandos aninhados, pode usar **`group_id`** (só faz efeito quando o path tem `/`); ver [Grupos de help](../technical-reference/plugins.md#grupos-de-help-groupsyaml-group_id-e-cobra).

### Folha com `entrypoint`

- **`command`** — obrigatório.
- **`entrypoint`** — ficheiro relativo à pasta do manifest; tem de **existir** e ficar **dentro** do plugin. Se terminar em **`.sh`**, o MB executa com **bash**; caso contrário trata como binário.

Pode combinar **`entrypoint`** raiz com **`flags`**: sem flag corre o entrypoint padrão; com flag corre o script dessa flag.

### Folha só com `flags` (flags-only)

Sem `entrypoint` no nível raiz. Lista **`flags`** (ver exemplo abaixo). Se o utilizador invocar o comando **sem nenhuma flag**, o CLI mostra o help e **não** executa script.

Exemplo mínimo flags-only:

```yaml
command: do
description: "Ações por flag"
flags:
  - name: deploy
    description: Faz o deploy
    entrypoint: deploy.sh
    commands:
      long: deploy
      short: d
```

Exemplo completo: [examples/plugins/tools/do](https://github.com/carlosdorneles-mb/mb-cli/tree/main/examples/plugins/tools/do).

## Campos opcionais (resumo)

### Texto e visibilidade

| Campo | Função |
|-------|--------|
| `description` | Short do Cobra (listagens e resumo do `--help`). |
| `long_description` | Long do Cobra (multi-linha; usar `\|` ou `>` no YAML). |
| `readme` | Ativa `--readme` / `-r` no comando folha (Markdown no terminal). |
| `hidden` | `true`: não aparece nos helps; comando continua invocável. |

### Uso, args e aliases (Cobra)

| Campo | Função |
|-------|--------|
| `use` | Sufixo da linha de uso (prefixado pelo nome do comando). Ex.: `"<name>"` obrigatório, `"[env]"` opcional. |
| `args` | Número de argumentos posicionais **obrigatórios** passados ao script (`0` = sem validação). |
| `aliases` | Lista de nomes alternativos para o mesmo comando. |
| `example` | Texto de exemplo no help. |
| `deprecated` | Mensagem ao **executar** (aviso de obsoleto; o comando ainda corre). |

Manifesto de exemplo com vários destes campos:

```yaml
command: meu-comando
description: "Descrição curta"
long_description: |
  Texto longo no help.
entrypoint: run.sh
use: "<name>"
args: 1
aliases:
  - x
example: |
  mb tools meu-comando foo
hidden: false
```

As **flags globais** (`-v`, `-q`, `--env-file`, `-e`) são sempre consumidas pelo CLI. O que o script recebe em `$1`, `$2`, … e o comportamento com flags desconhecidas: [Execução: flags e argumentos](../technical-reference/plugins.md#execução-flags-e-argumentos).

### `env_files` (opcional)

Ficheiros `.env` **dentro da pasta do plugin**, mesclados na execução conforme **`--env-group`**:

```yaml
env_files:
  - file: .env
  - file: .env.local
    group: local
```

Manifestos só de **categoria** ignoram `env_files`. Ordem de precedência do ambiente: [Variáveis de ambiente](./environment-variables.md).

### Grupos no help (`groups.yaml` e `group_id`)

Para secções personalizadas no help de comandos **aninhados** (ex. «INFRAESTRUTURA»):

1. Defina grupos em ficheiros **`groups.yaml`** (vários por pacote; registo **global** no sync, **último vence** se o mesmo `id` repetir).
2. Nos manifests de folhas ou categorias aninhadas, use **`group_id:`** com um `id` registado.

Comandos **top-level** sob `mb` ignoram `group_id` para secção (ficam em COMANDOS DE PLUGINS). Detalhes, regex de `id` e debug com **`mb -v`**: [Grupos de help](../technical-reference/plugins.md#grupos-de-help-groupsyaml-group_id-e-cobra).

## Script ou binário

```bash
#!/bin/sh
echo "Plugin rodando!"
```

Torne executável: `chmod +x run.sh`.

**Helpers MB** (após `mb plugins sync`): no shell, `. "$MB_HELPERS_PATH/all.sh"` ou `log.sh`. Lista: [Helpers de shell](../technical-reference/helpers-shell.md). **gum** é opcional nos scripts.

## Registar e sincronizar

### Um pacote com `manifest.yaml` na raiz

```bash
mb plugins add /caminho/para/meu-pacote --name meu-plugin
# ou, a partir da raiz do pacote:
mb plugins add .
```

O MB **não** valida o manifest na hora do `add` como no modo coleção; o **`mb plugins sync`** (disparado pelo add) pode mostrar **avisos** e **ignorar** manifests inválidos. Corrija avisos e volte a sincronizar.

### Vários plugins numa pasta (modo coleção)

A pasta **não** tem `manifest.yaml` na raiz. Cada **subdiretório direto** que tenha `manifest.yaml` é candidato; o MB valida cada um no add (entrada inválida → aviso e ignora). **Não** use **`--name`** se forem encontrados **vários** candidatos.

### Remoto

```bash
mb plugins add https://github.com/org/repo
mb plugins add https://github.com/org/repo --tag v1.0.0 --name meu-nome
```

### Manual em `PluginsDir`

```bash
mb plugins sync
mb plugins list
```

## Repositório com vários comandos

Um único `plugins add` cobre **toda a árvore**; o path no CLI **não** inclui o nome da instalação. Mais detalhes: [Repositório com vários plugins](./plugin-commands.md#repositório-com-vários-plugins).

## README opcional

Com `readme: README.md` na mesma pasta que o manifest da folha, o utilizador pode usar **`--readme`** para ver o Markdown no terminal.

---

Para comandos `mb plugins` (list, remove, update) e indicação **(local)** no help, veja [Comandos de plugins](./plugin-commands.md).
