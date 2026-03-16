---
sidebar_position: 2
---

# Criar um plugin

Este guia mostra o passo a passo para criar um plugin do MB CLI. Para uma visão técnica de como o CLI descobre e executa plugins, veja [Plugins (referência técnica)](./plugins.md).

## 1. Estrutura do diretório

Cada plugin fica em uma pasta. A hierarquia de pastas define a **categoria** no CLI. Exemplo: uma pasta `tools/meu-comando/` vira o comando `mb tools meu-comando`.

Você pode criar o plugin em qualquer lugar para desenvolvimento e depois instalá-lo de duas formas:

- **Remoto** — Publicar em um repositório Git e outras pessoas (ou você) instalam com `mb plugins add <url>`.
- **Local** — Registrar o path do diretório onde está desenvolvendo, sem copiar nada: `mb plugins add .` (diretório atual) ou `mb plugins add /caminho/para/meu-plugin`. Útil para testar enquanto desenvolve.

## 2. Manifesto `manifest.yaml`

Crie `manifest.yaml` na pasta raiz do plugin (ou em subpastas, se quiser categorias aninhadas):

```yaml
command: meu-comando   # nome do comando (opcional; padrão = nome da pasta)
description: "Descrição curta para o help"
type: sh               # sh = script shell; bin = executável
entrypoint: run.sh     # arquivo a executar (relativo à pasta do plugin)
readme: README.md      # opcional: flag --readme exibe com glow
```

Campos principais: `command`, `description`, `type`, `entrypoint`. O `readme` é opcional; se existir, o comando ganha a flag `--readme`.

Para plugins que só expõem flags (sem entrypoint único), use o campo `flags` no manifesto; consulte a [Referência de comandos](./reference.md) ou o código do MB para o formato.

## 3. Script ou binário

Para `type: sh`, crie o script referido em `entrypoint` (ex.: `run.sh`):

```bash
#!/bin/sh
echo "Plugin rodando!"
echo "Variável injetada: API_KEY=${API_KEY:-não definida}"
```

Torne o script executável (`chmod +x run.sh`). Para `type: bin`, use um executável compilado (Go, Rust, etc.) e indique-o em `entrypoint`.

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
