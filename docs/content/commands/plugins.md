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
| `--package <id>` | Identificador do pacote. Se omitido, usa o nome do repositório ou do diretório. Para regras de nomenclatura, veja [Nome do pacote](../technical-reference/plugins.md#nome-do-pacote-identificador). |
| `--tag <tag>` | Instalar tag específica (remoto) |

**Subdiretório de plugins:** o MB detecta automaticamente se os plugins estão em `src/` (ou o valor de `MB_PLUGIN_SUBDIR`). Para detalhes, veja [Subdiretório de plugins](../technical-reference/plugins.md#subdiretorio-de-plugins).

Após `add`, o sync é executado automaticamente.

**Local vs remoto:** plugins locais não são copiados — o path fica registado em `plugin_sources.local_path`. Remotos são clonados para `PluginsDir/<pacote>`.

### `mb plugins list`

Lista plugins instalados com informações detalhadas.

**Modos de exibição:**

- **Interativo (terminal):** Interface fzf com colunas simplificadas (PACOTE | COMANDO) e preview automático no lado direito com detalhes completos
- **Pipe/redirecionamento:** Tabela completa com colunas (PACOTE | COMANDO | DESCRIÇÃO | VERSÃO | ORIGEM | ATUALIZAR)
- **JSON:** Todos os dados em formato estruturado com `--json`

```bash
# Modo interativo (padrão)
mb plugins list
mb plugins ls
mb plugins l

# Com verificação de atualizações
mb plugins list --check-updates

# Saída JSON completa
mb plugins list --json

# Pipe com colunas completas
mb plugins list | cat
mb plugins list | grep local
mb plugins list | wc -l
```

| Flag | Descrição |
|---|---|
| `--check-updates` | Verifica se há atualização disponível para cada plugin remoto |
| `--json` / `-J` | Saída em formato JSON com todos os dados do plugin |

**Preview automático:**

No modo interativo, ao navegar com ↑↓ um preview aparece automaticamente no lado direito mostrando detalhes do plugin selecionado:

- Pacote e comando
- Descrição completa
- Versão e origem
- URL/path
- Referência (tag/branch)

O preview é renderizado com `gum format` para melhor legibilidade e atualiza em tempo real conforme você navega.

**Colunas no modo interativo:**

| Coluna | Descrição |
|---|---|
| PACOTE | Identificador do pacote (usado em `remove` e `update`) |
| COMANDO | Caminho do comando na árvore do CLI |

**Preview automático (modo interativo):**

Ao navegar com ↑↓, o preview mostra informações adicionais:

| Campo | Descrição |
|---|---|
| ORIGEM | `local` (path) ou `remoto` (Git) |
| VERSÃO | Versão ou commit do plugin |
| URL | Path local ou URL do repositório |
| REF | Tag ou branch (quando aplicável) |

**Colunas no modo pipe/redirecionamento:**

| Coluna | Descrição |
|---|---|
| PACOTE | Identificador do pacote |
| COMANDO | Caminho do comando |
| DESCRIÇÃO | Descrição curta do plugin (truncada a 47 caracteres) |
| VERSÃO | Versão ou commit do plugin |
| ORIGEM | `local` ou `remoto` |
| ATUALIZAR | `sim` se há atualização disponível (só com `--check-updates`) |

**Exemplos de uso com JSON:**

```bash
# Listar apenas plugins locais
mb plugins list --json | jq '.plugins[] | select(.origin == "local")'
mb plugins list -J | jq '.plugins[] | select(.origin == "local")'

# Contar plugins remotos
mb plugins list --json | jq '[.plugins[] | select(.origin == "remoto")] | length'

# Listar nomes de pacotes únicos
mb plugins list --json | jq -r '.plugins[].package' | sort -u

# Filtrar plugins com atualização disponível
mb plugins list --json --check-updates | jq '.plugins[] | select(.updateAvailable == true)'
```

**Exemplos de uso com pipe:**

```bash
# Buscar por descrição
mb plugins list | grep -i deploy

# Contar plugins
mb plugins list | tail -n +4 | wc -l

# Filtrar por origem
mb plugins list | grep remoto
```

A coluna **PACOTE** é o identificador usado em `mb plugins remove <pacote> [<pacote>...]` / `mb plugins remove --all` e `mb plugins update <pacote> [<pacote>...]` / `mb plugins update --all`.

### `mb plugins remove`

Remove um ou mais plugins instalados.

```bash
# Remove um único plugin
mb plugins remove meu-plugin

# Remove múltiplos plugins
mb plugins remove foo bar baz

# Remove todos os plugins instalados
mb plugins remove --all
```

| Flag | Descrição |
|---|---|
| `--all` | Remove todos os plugins instalados |

Pede confirmação antes de remover.

### `mb plugins update`

Atualiza um ou mais plugins remotos (Git), ou todos com `--all`.

```bash
mb plugins update meu-plugin
mb plugins update foo bar
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
