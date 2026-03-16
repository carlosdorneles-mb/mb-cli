---
sidebar_position: 1
---

# Introdução ao MB CLI

O **MB CLI** é um CLI em Go para orquestrar plugins com UX em laranja, descoberta dinâmica via cache SQLite e injeção segura de variáveis de ambiente.

## O que o MB faz

- **Comandos dinâmicos**: plugins viram comandos `mb <categoria> <comando>` automaticamente.
- **Cache SQLite**: após `mb self sync`, o CLI não precisa escanear o disco a cada execução.
- **Um manifesto por plugin**: cada plugin declara nome, categoria e, opcionalmente, subcategoria em `manifest.yaml` (ex.: `mb infra ci deploy`).
- **Ambiente controlado**: variáveis são mescladas (sistema → arquivo .env → `--env`) e injetadas só no processo do plugin.
- **Help e erros estilizados**: [Fang](https://github.com/charmbracelet/fang) estiliza o `--help` e as mensagens de erro com o tema padrão.

## Próximos passos

**Guia (uso do dia a dia)**

- [Começar](getting-started) — pré-requisitos, build e instalação
- [Criar um plugin](creating-plugins) — passo a passo com manifest e entrypoint
- [Comandos de plugins](comandos-plugins) — como descobrir e executar comandos de plugins
- [Flags globais](flags-globais) — `--verbose`, `--quiet`, `--env`
- [Variáveis de ambiente](variaveis-ambiente) — ordem de precedência e como usar

**Referência técnica**

- [Arquitetura](arquitetura) — visão de alto nível do CLI (Cobra, cache, execução)
- [Plugins](plugins) — como o CLI descobre, armazena e executa plugins
- [Referência de comandos](reference) — tabela de comandos e flags
