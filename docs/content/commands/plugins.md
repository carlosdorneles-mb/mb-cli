---
sidebar_position: 2
---

# `mb plugins`

Gerencia plugins instalados no CLI — adicionar, listar, remover, atualizar e sincronizar.

## Subcomandos

### `mb plugins add`

Instala um plugin a partir de uma URL Git ou de um diretório local.

```bash
# Local (path ou diretório atual)
mb plugins add /caminho/para/meu-pacote --package meu-plugin
mb plugins add .

# Remoto (Git)
mb plugins add https://github.com/org/repo
mb plugins add https://github.com/org/repo --tag v1.0.0 --package meu-pacote
```

| Flag | Descrição |
|---|---|
| `--package <id>` | Identificador do pacote |
| `--tag <tag>` | Instalar tag específica (remoto) |

Após `add`, o sync é executado automaticamente.

**Local vs remoto:** plugins locais não são copiados — o path fica registado em `plugin_sources.local_path`. Remotos são clonados para `PluginsDir/<pacote>`.

### `mb plugins list`

Lista plugins instalados com pacote, caminho do comando, descrição, versão, origem (local/remoto) e URL/path.

```bash
mb plugins list
```

### `mb plugins remove`

Remove um plugin instalado (diretório + registo no cache).

```bash
mb plugins remove meu-plugin
```

Pede confirmação antes de remover.

### `mb plugins update`

Atualiza plugins remotos (Git).

```bash
mb plugins update meu-plugin
mb plugins update --all
```

| Flag | Descrição |
|---|---|
| `--all` | Atualiza todos os plugins com URL Git |

Plugins locais não podem ser atualizados por este comando.

### `mb plugins sync`

Rescaneia o diretório de plugins e os paths locais registados, atualiza o cache SQLite e garante os helpers de shell.

```bash
mb plugins sync
```

Disparado automaticamente após `plugins add`. Obrigatório se editou ficheiros diretamente em `PluginsDir`.

## Onde ficam os plugins

| Plataforma | Caminho |
|---|---|
| **Linux** | `~/.config/mb/plugins` |
| **macOS** | `~/Library/Application Support/mb/plugins` |

Para detalhes sobre como criar plugins, veja [Criar um plugin](../plugin-authoring/create-a-plugin.md).
