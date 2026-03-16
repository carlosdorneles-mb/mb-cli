---
sidebar_position: 7
---

# Helpers de shell

O MB CLI injeta no ambiente dos plugins a variável **`MB_HELPERS_PATH`**, que aponta para um script que carrega funções de shell reutilizáveis. Assim você pode usar helpers como `log` sem duplicar código no seu plugin. O helper de log usa [gum](https://github.com/charmbracelet/gum) — instale-o para que `log` funcione.

## Como carregar

No início do script do plugin (por exemplo em `run.sh`), source o script:

```sh
#!/bin/sh
. "$MB_HELPERS_PATH"

# A partir daqui você pode usar os helpers
log info "Olá!"
```

O path em `MB_HELPERS_PATH` é o do arquivo `index.sh` em `~/.config/mb/lib/shell/` (criado na primeira execução de um plugin).

## Helpers disponíveis

### log

Log que respeita `MB_QUIET` e `MB_VERBOSE` (flags `-q` e `-v` do CLI). Usa `gum log -l` por baixo.

**Uso:** `log <level> <mensagem...>`

**Níveis:** `none`, `debug`, `info`, `warn`, `error`, `fatal`

**Comportamento:**

- **`MB_QUIET=1`** — Só exibe mensagens com nível `error` e `fatal`.
- **`MB_VERBOSE=1`** — Exibe todos os níveis, incluindo `debug`.
- **Caso contrário** — Exibe `info`, `warn`, `error`, `fatal`; o nível `debug` é omitido.

Exemplos:

```sh
log info "Processando..."
log debug "Detalhe interno: $var"
log warn "Aviso"
log error "Algo falhou"
```
