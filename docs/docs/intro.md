---
sidebar_position: 1
---

# Introdução ao MB CLI

O **MB CLI** é uma ferramenta de linha de comando para **estender o que você pode fazer com um único programa**: em vez de lembrar de vários scripts espalhados, você instala ou registra **plugins** e passa a invocá-los de forma uniforme — por exemplo `mb ferramentas meu-comando`, com ajuda integrada e o mesmo jeito de passar opções de ambiente.

## O que o MB faz por você

- **Um comando para muitas tarefas**: os plugins aparecem na árvore de comandos do `mb`; você descobre o que existe com `mb help` ou `mb plugins list`.
- **Sincronização simples**: depois de instalar ou alterar plugins, `mb plugins sync` atualiza a lista de comandos disponíveis — sem precisar reinstalar o programa principal.
- **Organização por pastas e manifesto**: cada plugin descreve-se num arquivo `manifest.yaml` (nome, descrição, como executar). A estrutura de pastas define categorias e subcomandos na linha de comando.
- **Ambiente sob controle**: variáveis de ambiente podem vir do sistema, de arquivo e da linha de comando (`--env`), e são aplicadas de forma previsível ao rodar um plugin.

## Próximos passos

**Guia (uso do dia a dia)**

- [Começar](./guide/getting-started.md) — pré-requisitos, instalação e primeiros passos
- [Comandos de plugins](./guide/plugin-commands.md) — como descobrir e executar comandos de plugins
- [Criar um plugin](./guide/creating-plugins.md) — passo a passo com manifest e scripts
- [Flags globais](./guide/global-flags.md) — `--verbose`, `--quiet`, `--env`
- [Variáveis de ambiente](./guide/environment-variables.md) — ordem de precedência e como usar
- [Segurança](./guide/security.md) — de onde vem o código que roda e boas práticas

**Referência técnica** (implementação, cache, fluxos internos)

- [Arquitetura](./technical-reference/architecture.md) — como o CLI monta comandos, cache e execução
- [Plugins](./technical-reference/plugins.md) — descoberta de plugins, sync e execução em detalhe
- [Referência de comandos](./technical-reference/reference.md) — tabela de comandos e flags
- [Versionamento e release](./technical-reference/versioning-and-release.md) — versões e publicação do projeto
