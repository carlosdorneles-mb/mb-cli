---
sidebar_position: 1
---

# Criar um plugin

Guia passo a passo para montar um pacote de plugins do MB CLI.

- **Referência técnica** (scanner, sync, grupos de help): [Plugins](../technical-reference/plugins.md)
- **Helpers de shell** (log, memória, k8s, etc.): [Helpers para plugins](./shell-helpers.md)
- **Exemplos no repositório**: [examples/plugins](https://github.com/carlosdorneles-mb/mb-cli/tree/main/examples/plugins)

## 1. Estrutura de pastas

Cada plugin é uma árvore de diretórios com ficheiros `manifest.yaml`. A estrutura de pastas define os subcomandos:

```
meu-pacote/
├── manifest.yaml              # Categoria raiz (ex.: mb tools)
├── vscode/
│   ├── manifest.yaml          # mb tools vscode (folha executável)
│   └── run.sh
└── podman/
    ├── manifest.yaml          # mb tools podman (folha executável)
    └── run.sh
```

- **Cada pasta** pode ter um `manifest.yaml`.
- **Pasta sem `entrypoint` nem `flags`** = categoria (subcomando intermédio).
- **Pasta com `entrypoint`** ou só `flags` = folha executável (comando final).
- O nome do subcomando vem do campo **`command`** do manifest; se omitido, usa o **nome da pasta**.

Exemplo: `tools/hello/manifest.yaml` com `command: hello` → **`mb tools hello`**.

## 2. `manifest.yaml` — Referência de campos

### Campos principais

| Campo | Tipo | Obrigatório? | Descrição |
|-------|------|--------------|-----------|
| `command` | string | **Sim** (folhas) | Nome do subcomando na CLI. Se omitido, usa o nome da pasta. |
| `entrypoint` | string | Sim (folha com script) | Ficheiro executável relativo à pasta do manifest. Se terminar em `.sh`, executa com **bash**; caso contrário, trata como binário. |
| `description` | string | Não | Texto curto no `--help` e listagens. |
| `long_description` | string | Não | Texto longo no `--help` (multi-linha; usar `\|` ou `>` no YAML). |
| `readme` | string | Não | Ficheiro Markdown na mesma pasta; ativa `--readme` / `-r` no comando. |
| `hidden` | bool | Não | `true`: omite do `mb help` (comando continua invocável). |

### Flags e argumentos

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `use` | string | Sufixo da linha de uso. Ex.: `"<name>"` (obrigatório), `"[env]"` (opcional). |
| `args` | int | Nº de argumentos posicionais **obrigatórios** (`0` = sem validação). |
| `aliases` | `[]string` | Nomes alternativos para o comando. Funciona em **folhas** e **categorias**. |
| `example` | string | Texto de exemplo no `--help`. |
| `deprecated` | string | Mensagem ao executar (aviso de obsoleto; o comando ainda corre). |

### Flags do comando (`flags`)

Quando a folha **não tem `entrypoint`** no nível raiz, use `flags` para definir sub-ações:

```yaml
command: do
description: "Ações por flag"
flags:
  - name: deploy
    description: Faz o deploy
    entrypoint: deploy.sh
    envs:
      - MODE=production
    commands:
      long: deploy
      short: d
```

| Campo (FlagEntry) | Tipo | Descrição |
|-------------------|------|-----------|
| `name` | string | Identificador da flag (obrigatório). |
| `description` | string | Texto no `--help` da flag. |
| `entrypoint` | string | Script executado quando a flag é usada. |
| `envs` | `[]string` | Pares `KEY=VALUE` injetados só quando a flag é ativa. |
| `commands.long` | string | Flag longa (ex.: `--deploy`). Se omitido, usa `name`. |
| `commands.short` | string | Flag curta (ex.: `-d`). Opcional. |

Se o utilizador invocar o comando **sem nenhuma flag**, o CLI mostra o help e **não** executa nenhum script.

> **Combinação `entrypoint` + `flags`**: pode ter ambos. Sem flag → corre o entrypoint raiz; com flag → corre o script dessa flag.

### `env_files` (opcional)

Ficheiros `.env` **dentro da pasta do plugin**, mesclados na execução conforme `--env-vault`:

```yaml
env_files:
  - file: .env
  - file: .env.local
    vault: local
```

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `file` | string | Path relativo à pasta do plugin. |
| `vault` | string | Nome do vault (default: `default` se omitido). |

Manifestos só de **categoria** ignoram `env_files`.

### `group_id` (grupos de help)

Para secções personalizadas no `--help` de comandos **aninhados** (ex.: «INFRAESTRUTURA»):

```yaml
command: deploy
group_id: infra
```

O `group_id` deve corresponder a um `id` definido num ficheiro **`groups.yaml`** do mesmo pacote. Comandos **top-level** (logo abaixo de `mb`) ignoram `group_id`. Veja [Grupos de help](../technical-reference/plugins.md#grupos-de-help-groupsyaml-group_id-e-cobra).

## 3. Exemplos práticos

### Categoria simples

```yaml
# infra/manifest.yaml
command: infra
description: Comandos de infraestrutura
```

### Folha com script

```yaml
# infra/deploy/manifest.yaml
command: deploy
description: Faz o deploy da aplicação
long_description: |
  Executa o pipeline de deploy no ambiente atual.
  Usa as variáveis DEPLOY_ENV e DEPLOY_TARGET.
entrypoint: deploy.sh
use: "[environment]"
args: 0
aliases:
  - d
example: |
  mb infra deploy production
  mb infra deploy staging --env DEPLOY_TARGET=us-east
```

```bash
#!/bin/bash
. "$MB_HELPERS_PATH/all.sh"

env="${1:-production}"
log info "Deploying to $env..."
# ... lógica do deploy
```

### Folha só com flags

```yaml
# tools/do/manifest.yaml
command: do
description: Ações utilitárias
flags:
  - name: install
    description: Instala a ferramenta
    entrypoint: install.sh
    commands:
      long: install
      short: i
  - name: update
    description: Atualiza a ferramenta
    entrypoint: update.sh
    commands:
      long: update
      short: u
  - name: uninstall
    description: Remove a ferramenta
    entrypoint: uninstall.sh
    commands:
      long: uninstall
      short: x
```

### Plugin com `env_files`

```yaml
# api/serve/manifest.yaml
command: serve
description: Inicia o servidor API
entrypoint: run.sh
env_files:
  - file: .env
  - file: .env.production
    vault: production
  - file: .env.staging
    vault: staging
```

## 4. Registar e testar

### Convenção de subdiretório

O MB CLI detecta automaticamente se os plugins estão num **subdiretório** do repositório.
Por padrão, procura em **`src/`**. Se não encontrar manifests nessa pasta, faz **fallback** para a raiz do pacote.

```
repo/
├── README.md
├── src/                     ← plugins aqui (detectado automaticamente)
│   ├── manifest.yaml
│   └── deploy/
│       ├── manifest.yaml
│       └── deploy.sh
└── docs/
    └── guia.md
```

Para **customizar** o subdiretório, defina a variável de ambiente **`MB_PLUGIN_SUBDIR`**:

```bash
# Usar "lib/" em vez de "src/"
MB_PLUGIN_SUBDIR=lib mb plugins add https://github.com/org/repo

# Desativar detecção automática (escanear sempre da raiz)
MB_PLUGIN_SUBDIR= mb plugins add https://github.com/org/repo
```

| `MB_PLUGIN_SUBDIR` | Comportamento |
|---------------------|---------------|
| não definida | Tenta `src/` → se vazio, raiz |
| `lib` | Tenta `lib/` → se vazio, raiz |
| `""` (vazio) | Escaneia direto da raiz |

### Nome do pacote

Se `--package` não for informado:

- **Remoto (Git):** usa o nome do repositório (último segmento da URL).
  Ex.: `https://github.com/org/infra-tools` → `infra-tools`
- **Local:** usa o nome do diretório.
  Ex.: `/caminho/para/meu-plugin` → `meu-plugin`
- **Coleção:** usa o nome de cada subdiretório.
  Ex.: `repo/foo/` + `repo/bar/` → `foo` e `bar`

Esse nome é o identificador que aparece na coluna **PACOTE** de `mb plugins list`
e é usado em `mb plugins remove <pacote>` e `mb plugins update <pacote>`.

### Local (sem copiar ficheiros)

```bash
mb plugins add /caminho/para/meu-pacote --package meu-plugin
# ou, a partir da raiz do pacote:
cd /caminho/para/meu-pacote
mb plugins add .
```

O sync é automático. Verifique:

```bash
mb plugins list
mb help                # ou mb help <categoria>
mb <categoria> <comando>
```

### Remoto (Git)

```bash
mb plugins add https://github.com/org/repo
mb plugins add https://github.com/org/repo --tag v1.0.0 --package meu-plugin
```

### Manual (copiar para o diretório de plugins)

```bash
# Linux
cp -r meu-pacote ~/.config/mb/plugins/

# macOS
cp -r meu-pacote ~/Library/Application\ Support/mb/plugins/

mb plugins sync
```

## 5. Escrever o script {#5-escrever-o-script}

### Receber argumentos

O script recebe argumentos posicionais em `$1`, `$2`, etc.:

```bash
#!/bin/bash
. "$MB_HELPERS_PATH/all.sh"

name="${1:-world}"
log info "Hello, $name!"
```

### Variáveis de ambiente

O plugin recebe variáveis mescladas (sistema, defaults, vault, `--env`). Veja [Variáveis de ambiente](../user-guide/environment-variables.md) para a ordem de precedência.

Além disso, variáveis de contexto são injetadas (`MB_CTX_*`): path do comando, comando pai, irmãos, filhos, etc. Veja [Contexto de invocação](../technical-reference/plugin-invocation-context.md).

### Usar helpers

```bash
#!/bin/bash
. "$MB_HELPERS_PATH/all.sh"

# Log
log info "Iniciando..."
log debug "Detalhe: $VAR"

# Detecção de OS
if is_mac; then
  brew install curl
elif is_linux_debian; then
  sudo apt-get install -y curl
fi

# Memória entre execuções
if ! mem_has "myplugin" "last_run"; then
  mem_set "myplugin" "last_run" "$(date)"
fi
```

Lista completa: [Helpers para plugins](./shell-helpers.md).

### Códigos de saída

Para plugins com `install.sh` / `update.sh` que podem precisar de `sudo`:

| Código | Significado |
|--------|-------------|
| **86** | Sem privilégio para gestores de pacote (não é falha no batch `--update-all`). |
| **87** | Ferramenta não instalada ao atualizar (ignorado no batch). |

Use `return "${MB_EXIT_UPDATE_SKIPPED_SUDO:-86}"` para manter compatibilidade:

```bash
warn_and_skip_without_sudo "Minha Ferramenta" || return $?
```

## 6. README opcional

Coloque um `README.md` na mesma pasta do manifest da folha:

```yaml
command: deploy
entrypoint: deploy.sh
readme: README.md
```

O utilizador pode ver o Markdown no terminal:

```bash
mb infra deploy --readme
```

## Resumo rápido

| Passo | O quê |
|-------|-------|
| 1 | Criar pastas + `manifest.yaml` em cada nível |
| 2 | Folha: `command` + `entrypoint` (ou só `flags`) |
| 3 | `mb plugins add <path ou URL> [--package nome]` |
| 4 | `mb plugins sync` (automático após `add`) |
| 5 | Testar: `mb plugins list`, `mb help`, `mb <cat> <cmd>` |

## Próximos passos

- [Helpers de shell](./shell-helpers.md) — log, memória, OS, k8s, snap, homebrew…
- [Comandos de plugins](../user-guide/plugin-commands.md) — `mb plugins list`, `remove`, `update`
- [Variáveis de ambiente](../user-guide/environment-variables.md) — vaults, secrets, precedência
- [Plugins (referência técnica)](../technical-reference/plugins.md) — scanner, cache, sync em detalhe
