---
sidebar_position: 1
---

# Introdução ao MB CLI

O **MB CLI** é uma ferramenta de linha de comando que transforma **plugins em comandos dinâmicos**, com cache SQLite, injeção segura de variáveis de ambiente e helpers de shell poderosos.

> Em vez de manter vários scripts espalhados, você instala ou registra **plugins** e passa a invocá-los de forma uniforme: `mb <categoria> <comando>`, com ajuda integrada e controle total do ambiente de execução.

## Por que MB CLI?

| Problema | Solução MB CLI |
|----------|----------------|
| Scripts soltos sem padrão | Plugins com `manifest.yaml` e estrutura consistente |
| Variáveis de ambiente espalhadas | Mescla previsível: sistema → defaults → `--env` → `env_files` |
| Descoberta difícil de comandos | `mb help` e `mb plugins list` mostram tudo disponível |
| Reinstalar para adicionar funcionalidade | `mb plugins sync` atualiza comandos sem reinstalar o CLI |
| Sem reutilização entre projetos | Helpers de shell embutidos: log, memória, k8s, Homebrew, Flatpak… |

## Como funciona

```
┌─────────────────────────────────────────────────────────────┐
│  Plugins (Git ou local)                                     │
│  ├── tools/manifest.yaml                                    │
│  ├── infra/manifest.yaml                                    │
│  └── deploy/manifest.yaml                                   │
└──────────────────────┬──────────────────────────────────────┘
                       │ mb plugins sync
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Cache SQLite (cache.db)                                    │
│  → Comandos, flags, entrypoints, hashes                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Comandos dinâmicos na árvore do CLI                        │
│  mb tools vscode                                            │
│  mb infra ci                                                │
│  mb deploy production --env staging                         │
└─────────────────────────────────────────────────────────────┘
```

## O que o MB faz por você

- **Um comando para muitas tarefas** — plugins aparecem na árvore de comandos do `mb`; descubra o que existe com `mb help` ou `mb plugins list`.
- **Sincronização simples** — `mb plugins sync` atualiza comandos disponíveis sem reinstalar o programa principal.
- **Organização por pastas e manifesto** — cada plugin se descreve num `manifest.yaml` (nome, descrição, como executar). A estrutura de pastas define categorias e subcomandos.
- **Ambiente sob controle** — variáveis do sistema, de arquivos e da linha de comando (`--env`) são aplicadas de forma previsível e isolada por plugin.
- **Helpers de shell prontos** — bibliotecas de funções para log, memória, Kubernetes, Homebrew, Flatpak, GitHub e mais, disponíveis via `$MB_HELPERS_PATH`.

## Quick start

```bash
# Instalar o CLI
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash

# Adicionar um plugin
mb plugins add https://github.com/org/repo

# Sincronizar comandos
mb plugins sync

# Usar
mb <categoria> <comando> [flags]
```

## Estrutura da documentação

**Guia (uso do dia a dia)**

- [Começar](./getting-started/) — pré-requisitos, instalação e primeiros passos
- [Variáveis de ambiente](./user-guide/environment-variables.md) — ordem de precedência e como usar
- [Comandos de plugins](./user-guide/plugin-commands.md) — como descobrir e executar comandos
- [Flags globais](./user-guide/global-flags.md) — `--verbose`, `--quiet`, `--env`
- [Segurança](./user-guide/security.md) — de onde vem o código que roda e boas práticas

**Comandos do CLI** (referência de uso)

- [`mb envs`](./commands/envs.md) — listar, definir e remover variáveis de ambiente
- [`mb plugins`](./commands/plugins.md) — adicionar, listar, remover, atualizar e sincronizar
- [`mb run`](./commands/run.md) — executar qualquer programa com o ambiente mesclado
- [`mb update`](./commands/update.md) — atualizar plugins, CLI, ferramentas e sistema
- [`mb completion`](./commands/completion.md) — instalar e gerenciar autocompletar
- [`mb help`](./commands/help.md) — ajuda sobre qualquer comando

**Referência técnica** (implementação, cache, fluxos internos)

- [Arquitetura](./technical-reference/architecture.md) — como o CLI monta comandos, cache e execução
- [Plugins](./technical-reference/plugins.md) — descoberta, sync e execução em detalhe
- [Contexto de invocação](./technical-reference/plugin-invocation-context.md) — variáveis `MB_CTX_*` injetadas no plugin
- [Configuração do CLI](./technical-reference/cli-config.md) — `config.yaml` e configurações globais
- [Criar um plugin](./plugin-authoring/create-a-plugin.md) — passo a passo com manifest e scripts
- [Helpers de shell](./plugin-authoring/shell-helpers.md) — funções utilitárias para plugins
- [Referência de comandos](./technical-reference/reference.md) — tabela de comandos e flags
- [Versionamento e release](./technical-reference/versioning-and-release.md) — versões e publicação
