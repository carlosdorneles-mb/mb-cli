---
sidebar_position: 4
---

# Criar um plugin

## 1. Criar o diretório do plugin

No diretório de plugins do MB (ex.: `~/.config/mb/plugins`):

```bash
# Linux
mkdir -p ~/.config/mb/plugins/meu-plugin

# macOS
mkdir -p ~/Library/Application\ Support/mb/plugins/meu-plugin
```

## 2. Criar o manifesto `manifest.yaml`

Crie `manifest.yaml` dentro da pasta do plugin:

```yaml
name: meu-plugin      # nome do comando (ex.: mb tools meu-plugin)
category: tools       # categoria = subcomando pai
type: sh              # sh = script shell; bin = executável
entrypoint: run.sh    # arquivo a executar (relativo à pasta do plugin)
readme: README.md     # opcional: usado pela flag --readme (glow)
```

Campos obrigatórios: `name`, `category`, `type`, `entrypoint`.  
`readme` é opcional; se existir, o comando ganha a flag `--readme` para exibir o README com glow.

## 3. Criar o script ou binário

Para `type: sh`, crie o script referido em `entrypoint` (ex.: `run.sh`):

```bash
#!/bin/sh
echo "Plugin rodando!"
echo "Variável injetada: API_KEY=${API_KEY:-não definida}"
```

Torne o script executável:

```bash
chmod +x ~/.config/mb/plugins/meu-plugin/run.sh
```

Para `type: bin`, use um executável compilado (Go, Rust, etc.) e indique-o em `entrypoint`.

## 4. (Opcional) README para ajuda

Se você declarou `readme: README.md`, crie esse arquivo na mesma pasta. Ao rodar `mb tools meu-plugin --readme`, o MB renderiza esse Markdown com glow.

## 5. Registrar no cache e rodar

```bash
mb self sync
mb self list                    # lista plugins (incluindo meu-plugin)
mb tools meu-plugin             # executa o plugin
mb --env API_KEY=xyz tools meu-plugin   # com variável injetada
```
