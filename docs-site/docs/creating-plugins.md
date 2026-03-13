---
sidebar_position: 4
---

# Criar um plugin

Para uma visão completa de como os plugins funcionam e como criá-los, consulte a página **[Plugins](./plugins.md)**.

Ela inclui:

- Como funciona o diretório de plugins, o cache e a árvore de comandos
- Comandos `mb plugins add`, `list`, `remove`, `update`
- Passo a passo para criar um plugin: estrutura de pastas, `manifest.yaml` (`command`, `description`, `type`, `entrypoint`, `readme`), script ou binário, e como registrar com `mb self sync` ou instalar via `mb plugins add <url>`

Resumo rápido:

1. Crie uma pasta em `~/.config/mb/plugins/<categoria>/<comando>/` (ex.: `tools/meu-plugin/`).
2. Adicione `manifest.yaml` com `command`, `description`, `type: sh`, `entrypoint: run.sh`.
3. Crie o script `run.sh` (executável) e rode `mb self sync`.
4. Execute com `mb tools meu-plugin`.

Detalhes e exemplos completos estão em [Plugins](./plugins.md).
