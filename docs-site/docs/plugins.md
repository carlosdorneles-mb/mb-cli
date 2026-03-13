---
sidebar_position: 3
---

# Plugins

Esta página descreve como os plugins funcionam no MB CLI, como gerenciá-los (`mb plugins`) e como criar um plugin.

## Como funciona

### Diretório de plugins

O MB usa um único diretório para plugins, derivado de `os.UserConfigDir()`:

- **Linux**: `~/.config/mb/plugins`
- **macOS**: `~/Library/Application Support/mb/plugins`

### Descoberta e cache

O comando `mb self sync` percorre esse diretório em busca de arquivos `manifest.yaml`. Cada manifesto define um plugin (comando, descrição, tipo, script ou binário). Os plugins encontrados são gravados em `~/.config/mb/cache.db` (ou equivalente no macOS). Esse banco é consultado na inicialização do CLI para montar a árvore de comandos. Por isso é importante rodar `mb self sync` depois de adicionar ou alterar plugins manualmente — ou use `mb plugins add <url>` para instalar a partir de um repositório Git, que já dispara o sync ao final.

### Árvore de comandos

O MB agrupa plugins por **categoria** (estrutura de pastas). Cada pasta vira um subcomando, e cada plugin vira um subcomando dessa categoria:

- Plugin em `infra/ci/deploy/` → `mb infra ci deploy`
- Plugin em `tools/hello/` → `mb tools hello`

### Execução

Ao rodar um comando de plugin, o MB localiza o executável/script no cache, monta o ambiente (merge de env) e executa o processo. Você pode usar scripts shell (`type: sh`) ou binários (`type: bin`).

### Variáveis de ambiente

Antes de executar um plugin, o MB mescla:

1. Variáveis do sistema (`os.Environ()`)
2. Arquivo de defaults: `<UserConfigDir>/mb/env.defaults` (e opcionalmente `--env-file`)
3. Variáveis da linha de comando: `--env KEY=VALUE`

A **maior precedência** é a de `--env`. Para gerenciar defaults: `mb self env list`, `mb self env set KEY [VALUE]`, `mb self env unset KEY`.

---

## Comandos `mb plugins`

O MB oferece o comando `mb plugins` para instalar, listar, remover e atualizar plugins a partir de repositórios Git (GitHub, Bitbucket, GitLab).

### `mb plugins add <git-url>`

Instala um plugin a partir de uma URL Git.

- **Sem flags**: instala a **tag mais recente** do repositório (ordenação semver). Se o repositório não tiver tags, clona a branch padrão (ex.: `main`).
- **`--name <name>`**: usa `<name>` como nome do plugin (diretório de instalação). Se não informado, usa o nome do repositório (último segmento da URL).
- **`--tag <tag>`**: instala uma tag específica (ex.: `v1.2.0`).

Exemplos:

```bash
mb plugins add https://github.com/org/repo
mb plugins add https://github.com/org/repo --name meu-plugin
mb plugins add https://github.com/org/repo --tag v1.0.0
```

Após a instalação, o cache é atualizado automaticamente (`mb self sync` é executado).

### `mb plugins list [--check-updates]`

Lista os plugins instalados, com nome, comando (path), descrição, versão e URL de origem.

- **`--check-updates`**: verifica se há atualização disponível para cada plugin (consulta remoto e compara tag ou branch). Pode demorar um pouco.

Exemplo:

```bash
mb plugins list
mb plugins list --check-updates
```

### `mb plugins remove <name>`

Remove um plugin instalado. O `<name>` é o nome do plugin (diretório de instalação), o mesmo exibido na coluna NAME de `mb plugins list`. O comando pede confirmação antes de apagar o diretório e atualizar o cache.

Exemplo:

```bash
mb plugins remove meu-plugin
```

### `mb plugins update [name | --all]`

Atualiza um plugin ou todos.

- **`mb plugins update <name>`**: atualiza apenas o plugin indicado.
  - Se o plugin foi instalado por **tag**: busca novas tags no remoto e, se existir uma tag mais recente (semver), faz checkout dessa tag.
  - Se o plugin foi instalado por **branch**: faz `git pull` na branch (ex.: `main`) e atualiza a versão (SHA curto) no registry.
- **`mb plugins update --all`**: percorre todos os plugins com URL Git e aplica a mesma lógica de atualização.

Exemplos:

```bash
mb plugins update meu-plugin
mb plugins update --all
```

---

## Como criar um plugin

### 1. Estrutura do diretório

Cada plugin fica em uma pasta dentro do diretório de plugins. A hierarquia de pastas define a **categoria** no CLI. Exemplo: `~/.config/mb/plugins/tools/meu-comando/` vira o comando `mb tools meu-comando`.

### 2. Manifesto `manifest.yaml`

Crie `manifest.yaml` na pasta do plugin:

```yaml
command: meu-comando   # nome do comando (opcional; padrão = nome da pasta)
description: "Descrição curta para o help"
type: sh               # sh = script shell; bin = executável
entrypoint: run.sh     # arquivo a executar (relativo à pasta do plugin)
readme: README.md      # opcional: flag --readme exibe com glow
```

Campos principais: `command`, `description`, `type`, `entrypoint`. O `readme` é opcional; se existir, o comando ganha a flag `--readme`.

Para plugins que só expõem flags (sem entrypoint único), use o campo `flags` no manifesto; consulte a documentação do MB para o formato.

### 3. Script ou binário

Para `type: sh`, crie o script referido em `entrypoint` (ex.: `run.sh`):

```bash
#!/bin/sh
echo "Plugin rodando!"
echo "Variável injetada: API_KEY=${API_KEY:-não definida}"
```

Torne o script executável (`chmod +x run.sh`). Para `type: bin`, use um executável compilado (Go, Rust, etc.) e indique-o em `entrypoint`.

### 4. (Opcional) README

Se você declarou `readme: README.md`, crie esse arquivo na mesma pasta. Ao rodar `mb tools meu-comando --readme`, o MB renderiza o Markdown com glow.

### 5. Registrar e rodar

Se você criou o plugin **manualmente** no diretório de plugins:

```bash
mb self sync
mb plugins list                 # confira se aparece
mb tools meu-comando            # executa
mb --env API_KEY=xyz tools meu-comando
```

Se o plugin está em um **repositório Git**, outras pessoas podem instalar com:

```bash
mb plugins add https://github.com/sua-org/meu-plugin
```
