---
sidebar_position: 2
---

# Criar um plugin

Este guia mostra o passo a passo para criar um plugin do MB CLI. Para uma visão técnica de como o CLI descobre e executa plugins, veja [Plugins (referência técnica)](../technical-reference/plugins.md).

Há **plugins de exemplo** no repositório: [examples/plugins](https://github.com/carlosdorneles-mb/mb-cli/tree/main/examples/plugins). Use-os como referência ou registre com `make install-examples` na raiz do repo e depois `mb self sync`.

## 1. Estrutura do diretório

Cada plugin fica em uma pasta. A hierarquia de pastas define a **categoria** no CLI. Exemplo: uma pasta `tools/meu-comando/` vira o comando `mb tools meu-comando`.

Você pode criar o plugin em qualquer lugar para desenvolvimento e depois instalá-lo de duas formas:

- **Remoto** — Publicar em um repositório Git e outras pessoas (ou você) instalam com `mb plugins add <url>`.
- **Local** — Registrar o path do diretório onde está desenvolvendo, sem copiar nada: `mb plugins add .` (diretório atual) ou `mb plugins add /caminho/para/meu-plugin`. Útil para testar enquanto desenvolve.

## 2. Manifesto `manifest.yaml`

Crie `manifest.yaml` na pasta raiz do plugin (ou em subpastas, se quiser categorias aninhadas):

```yaml
command: meu-comando   # obrigatório quando há entrypoint ou flags
description: "Descrição curta para o help"
entrypoint: run.sh     # script ou binário (relativo à pasta do plugin); tipo inferido pelo sufixo
readme: README.md      # opcional: flag --readme exibe com glow
hidden: false          # opcional: omite do help global (comando ainda funciona)
```

#### `description` (opcional)

Descrição **curta** do comando, exibida na listagem e no resumo do `--help` (Cobra **Short**). Se omitida, o CLI usa algo como "Executa &lt;caminho&gt;".

#### `long_description` (opcional)

Texto **longo** exibido no help do comando (Cobra **Long**). Pode ter várias linhas no YAML (use `|` ou `>`). Útil para explicar uso, exemplos ou requisitos.

#### `command` (obrigatório para comandos executáveis)

Nome do comando no CLI. **É obrigatório** quando o manifest define um plugin executável (tem `entrypoint` ou `flags`). Se omitido nesses casos, o plugin é **ignorado** no scan e um aviso de validação é exibido ("command é obrigatório quando há entrypoint ou flags"). Para pastas que são só categoria (sem entrypoint e sem flags), o nome da pasta continua sendo usado quando `command` não é informado.

#### `entrypoint` (para comando “folha” executável)

Caminho do **arquivo a rodar**, relativo à pasta onde está o `manifest.yaml`. Ex.: `run.sh`, `bin/meu-plugin`. O MB resolve o path de forma absoluta na execução. O **tipo de execução** é inferido pelo sufixo: se terminar em **`.sh`**, executa com **bash** (script); caso contrário, como binário. Não é necessário declarar `type` no manifesto.

É possível definir **entrypoint** e **flags** no mesmo manifest: ao executar o comando sem flag (ex.: `mb tools do`), o MB roda o entrypoint padrão; ao passar uma flag (ex.: `mb tools do --deploy`), roda o script daquela flag. Ex.: com `entrypoint: run.sh` e uma flag `deploy` com `entrypoint: deploy.sh`, `mb tools do` executa run.sh e `mb tools do --deploy` executa deploy.sh.

Para plugins que **só expõem flags** (sem execução padrão), não use `entrypoint` no nível raiz; use apenas o campo `flags` como **lista**. Cada item tem `name`, `description` (exibida no `--help` da flag), `entrypoint` e `commands` com `long` (nome da flag, ex.: `--deploy`) e `short` (opcional, uma letra, ex.: `-d`). O `short` deve ser único entre as flags do comando. Ex.:

```yaml
command: do
description: "Ações por flag (deploy, rollback)"
flags:
  - name: deploy
    description: Faz o deploy
    entrypoint: deploy.sh
    commands:
      long: deploy
      short: d
  - name: rollback
    description: Reverte o último deploy
    entrypoint: rollback.sh
    commands:
      long: rollback
      short: r
```

O usuário executa o comando passando a flag desejada: **`mb tools do --deploy`** ou **`mb tools do -d`** rodam `deploy.sh`; **`mb tools do --rollback`** ou **`mb tools do -r`** rodam `rollback.sh`. As descrições aparecem ao rodar `mb tools do --help`. Se rodar sem nenhuma flag (`mb tools do`), o CLI exibe o help e não executa script. Há um exemplo completo em [examples/plugins/tools/do](https://github.com/carlosdorneles-mb/mb-cli/tree/main/examples/plugins/tools/do).

Detalhes em [Plugins (referência técnica)](../technical-reference/plugins.md#execução-flags-e-argumentos-passados-ao-plugin).

#### `env_files` (opcional)

Lista de arquivos `.env` **dentro da pasta do plugin** carregados na execução, conforme o **`--env-group`** global do MB:

- **`file`** — Caminho relativo ao `manifest.yaml` (ex.: `.env`, `.env.local`).
- **`group`** — Opcional. Se omitido, vale o grupo lógico **`default`** (mesmo conjunto usado quando você **não** passa `--env-group`). Com **`--env-group local`**, só entram linhas cujo `group` é `local`.

Exemplo:

```yaml
env_files:
  - file: .env
  - file: .env.local
    group: local
```

Assim, `mb --env-group local … seu-comando` mescla as variáveis de `.env` daquele plugin; sem `--env-group`, `.env.local` (grupo default) entra na pilha. O nome do grupo segue as mesmas regras de `mb self env --group`. Se o arquivo declarado não existir na hora de rodar, o comando falha com mensagem clara. Manifestos que são **só categoria** (sem `entrypoint` nem `flags`) ignoram `env_files`.

Ordem de precedência completa: veja [Variáveis de ambiente](environment-variables.md).

#### Uso, argumentos e ajuda (Cobra)

Você pode customizar a linha de uso, a quantidade de argumentos e o help do comando com estes campos opcionais:

| Campo | Descrição |
|-------|-----------|
| **`use`** | String da **linha de uso** no help (sufixo). O valor informado é **sempre prefixado pelo nome do comando** na linha de uso do Cobra. **Convenção:** use **`<nome>`** para argumento **obrigatório** e **`[nome]`** para **opcional**. Ex.: `command: meu-comando` e `use: "<name>"` resultam em `meu-comando <name>` no help; `use: "[env]"` → argumento opcional. Vários podem ser combinados, ex.: `use: "<name> [options]"`. |
| **`args`** | Número **inteiro** de argumentos posicionais **obrigatórios**. Ex.: `args: 1` faz com que `mb tools meu-comando do` passe "do" como primeiro argumento ao script (e não como subcomando). Se omitido ou 0, não há validação de quantidade. |
| **`aliases`** | Lista de **strings**: nomes alternativos para invocar o mesmo comando. Ex.: `aliases: ["x", "run"]` permite `mb tools x` ou `mb tools run` em vez de `mb tools meu-comando`. |
| **`example`** | String exibida como **exemplo** no help do comando. Ex.: `example: "mb tools meu-comando do"`. |
| **`deprecated`** | Mensagem exibida quando o comando for **executado** (aviso de obsoleto). Ex.: `deprecated: "Use 'mb tools novo-comando' em vez disso."` O CLI mostra o aviso em português ("Comando \"&lt;nome&gt;\" está obsoleto: &lt;sua mensagem&gt;") e ainda executa o plugin. |
| **`hidden`** | Se `true`, o comando (ou a **categoria**, em manifest só de categoria) **não aparece** no `mb --help` e nos helps dos pais, mas continua **invocável** (ex.: `mb tools meu-comando`). Útil para comandos internos ou experimentais. O **autocompletar do shell** pode ainda sugerir o nome, conforme a configuração do completion. |

Exemplo de manifest com esses campos:

```yaml
command: meu-comando
description: "Descrição curta"
long_description: |
  Texto longo no help.
  Pode ter várias linhas.
entrypoint: run.sh
use: "<name>"
args: 1
aliases:
  - x
example: |
  "mb tools meu-comando"
  "mb tools x"
deprecated: ""   # deixe vazio ou omita se não for obsoleto
hidden: false
```

Com `use: "<name>"` e `args: 1`, invocações como **`mb tools meu-comando postman`** passam "postman" como primeiro argumento ao script. As **flags globais** (`-v`, `-q`, `--env-file`, `-e`) são sempre consumidas pelo CLI e não chegam ao script; `--help`/`-h` exibe o help do comando e não é repassado; os demais argumentos posicionais são repassados ao entrypoint. Detalhes em [Plugins (referência técnica)](../technical-reference/plugins.md#execução-flags-e-argumentos-passados-ao-plugin).

## 3. Script ou binário

Se o entrypoint termina em **`.sh`**, crie o script nesse caminho (ex.: `run.sh`):

```bash
#!/bin/sh
echo "Plugin rodando!"
echo "Variável injetada: API_KEY=${API_KEY:-não definida}"
```

Torne o script executável (`chmod +x run.sh`). Se o entrypoint não terminar em `.sh`, o MB trata como binário e executa o arquivo diretamente (ex.: executável Go, Rust, C).

### Usando os helpers do MB

Os helpers são instalados quando você roda **`mb self sync`** (ou ao adicionar um plugin com `mb plugins add`). Se o plugin for shell, você pode importar os helpers em `$MB_HELPERS_PATH` (diretório) para ter acesso a funções como `log`, que respeitam `MB_QUIET` e `MB_VERBOSE`. No início do script:

```sh
. "$MB_HELPERS_PATH/all.sh"
log info "Processando..."
```

Para carregar só o helper de log: `. "$MB_HELPERS_PATH/log.sh"`. Veja [Helpers de shell](../technical-reference/helpers-shell.md) para a lista de helpers e [Flags globais](./global-flags.md) para o efeito de `-v` e `-q`.

Além dos helpers, você pode usar os comandos do [gum](https://github.com/charmbracelet/gum) nos scripts do plugin (ex.: `gum choose`, `gum input`, `gum confirm`, `gum filter`) para criar interfaces interativas. O gum é opcional; se estiver instalado no sistema, os scripts podem chamá-lo normalmente.

## 4. (Opcional) README

Se você declarou `readme: README.md`, crie esse arquivo na mesma pasta. Ao rodar `mb tools meu-comando --readme`, o MB renderiza o Markdown no terminal (com glow, se instalado).

## 5. Registrar e rodar

### Desenvolvimento local (path ou diretório atual)

No diretório do plugin (ou de um nível acima), rode:

```bash
mb plugins add . --name meu-plugin
# ou, de qualquer lugar:
mb plugins add /caminho/para/meu-plugin --name meu-plugin
```

O CLI valida se o diretório contém pelo menos um `manifest.yaml` e registra o path. Nada é copiado para a pasta de plugins. Depois:

```bash
mb plugins list    # confira: ORIGEM = local
mb tools meu-comando
```

### Instalação a partir de um repositório Git (remoto)

Se o plugin está em um repositório, você ou outras pessoas podem instalar com:

```bash
mb plugins add https://github.com/sua-org/meu-plugin
```

O CLI clona o repositório para um subdiretório em `~/.config/mb/plugins/` e atualiza o cache. Use **`--name`** só para definir o **nome da instalação** (usado em `mb plugins list`, `mb plugins remove` e no path do clone); **não** altera o caminho dos comandos no CLI. Use `--tag` para uma tag específica.

### Repositório com vários plugins

Um único `mb plugins add <url>` ou `mb plugins add <path>` registra **toda a árvore** a partir da raiz do repositório ou do path. O caminho de cada comando no CLI (`mb …`) é montado **só** a partir dessa árvore:

- Em cada nível de pasta, se existir `manifest.yaml` com **`command`** preenchido, esse valor vira um segmento do path; senão usa-se o **nome da pasta**.
- Na pasta do comando executável (folha), o último segmento do path interno é sempre o **nome da pasta**; o nome do subcomando Cobra vem do `command` do manifest (obrigatório quando há `entrypoint` ou `flags`).

Exemplo: repositório com `tools/postman/manifest.yaml` e `dev/kinfo/manifest.yaml` gera **`mb tools postman`** e **`mb dev kinfo`** — o mesmo vale para clone local com `mb plugins add .` na raiz desse repositório.

Se duas fontes (dois `plugins add`) expuserem o **mesmo** caminho de comando, o `mb self sync` falha com mensagem de conflito até você remover ou ajustar uma das fontes.

### Plugin criado manualmente no diretório de plugins

Cada clone ou cópia fica em `~/.config/mb/plugins/<nome>/`. Os comandos seguem a estrutura **dentro** dessa pasta (ex.: `<nome>/tools/meu-comando/` → `mb tools meu-comando`). Se você copiou ou criou arquivos diretamente sob `~/.config/mb/plugins/<nome>/`:

```bash
mb self sync
mb plugins list
mb tools meu-comando
```

Para mais detalhes sobre os comandos `mb plugins` e sobre comandos de plugins no dia a dia, veja [Comandos de plugins](./plugin-commands.md).
