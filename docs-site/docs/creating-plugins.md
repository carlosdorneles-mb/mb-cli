---
sidebar_position: 2
---

# Criar um plugin

Este guia mostra o passo a passo para criar um plugin do MB CLI. Para uma visão técnica de como o CLI descobre e executa plugins, veja [Plugins (referência técnica)](./plugins.md).

Há **plugins de exemplo** no repositório: [examples/plugins](https://github.com/carlosdorneles-mb/mb-cli/tree/main/examples/plugins). Use-os como referência ou registre com `make install-examples` na raiz do repo e depois `mb self sync`.

## 1. Estrutura do diretório

Cada plugin fica em uma pasta. A hierarquia de pastas define a **categoria** no CLI. Exemplo: uma pasta `tools/meu-comando/` vira o comando `mb tools meu-comando`.

Você pode criar o plugin em qualquer lugar para desenvolvimento e depois instalá-lo de duas formas:

- **Remoto** — Publicar em um repositório Git e outras pessoas (ou você) instalam com `mb plugins add <url>`.
- **Local** — Registrar o path do diretório onde está desenvolvendo, sem copiar nada: `mb plugins add .` (diretório atual) ou `mb plugins add /caminho/para/meu-plugin`. Útil para testar enquanto desenvolve.

## 2. Manifesto `manifest.yaml`

Crie `manifest.yaml` na pasta raiz do plugin (ou em subpastas, se quiser categorias aninhadas):

```yaml
command: meu-comando   # opcional; se omitido = nome da pasta
description: "Descrição curta para o help"
entrypoint: run.sh     # script ou binário (relativo à pasta do plugin); tipo inferido pelo sufixo
readme: README.md      # opcional: flag --readme exibe com glow
```

#### `command` (opcional)

Nome do comando no CLI. Se omitido, o MB usa o **nome da pasta**. Ex.: pasta `meu-comando` → comando `mb tools meu-comando`. Útil quando você quer um nome diferente da pasta (ex.: pasta `deploy` em `infra/ci/deploy` continua como comando `deploy`).

#### `entrypoint` (para comando “folha” executável)

Caminho do **arquivo a rodar**, relativo à pasta onde está o `manifest.yaml`. Ex.: `run.sh`, `bin/meu-plugin`. O MB resolve o path de forma absoluta na execução. O **tipo de execução** é inferido pelo sufixo: se terminar em **`.sh`**, executa como script shell (`/bin/sh` + script); caso contrário, como binário. Não é necessário declarar `type` no manifesto.

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

Detalhes em [Plugins (referência técnica)](./plugins.md#execução-flags-e-argumentos-passados-ao-plugin).

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

Para carregar só o helper de log: `. "$MB_HELPERS_PATH/log.sh"`. Veja [Helpers de shell](./helpers-shell.md) para a lista de helpers e [Flags globais](./flags-globais.md) para o efeito de `-v` e `-q`.

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

O CLI clona o repositório para o diretório de plugins e atualiza o cache. Use `--name` para escolher o nome do plugin e `--tag` para uma tag específica.

### Plugin criado manualmente no diretório de plugins

Se você copiou ou criou o plugin diretamente em `~/.config/mb/plugins/<categoria>/<comando>/`:

```bash
mb self sync
mb plugins list
mb tools meu-comando
```

Para mais detalhes sobre os comandos `mb plugins` e sobre comandos de plugins no dia a dia, veja [Comandos de plugins](./comandos-plugins.md).
