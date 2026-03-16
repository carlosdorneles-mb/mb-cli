# Como executar o MB CLI localmente

Guia com os comandos exatos para rodar o CLI no seu ambiente, sem instalar.

---

## 1. Executar localmente usando seus plugins instalados

Use quando quiser rodar o código atual **e** que o CLI leia os plugins que você já instalou (em `~/.config/mb/plugins` no Linux ou `~/Library/Application Support/mb/plugins` no macOS). Não é definido `XDG_CONFIG_HOME`, então o MB usa o diretório de config real do sistema.

**Comando direto:**

```bash
go run . --help
go run . plugins list
go run . self sync
go run . <categoria> <comando>
```

**Via Makefile:**

```bash
make run-local
make run-local plugins list
make run-local self sync
make run-local tools meu-plugin
# Ou com ARGS: make run-local ARGS="self sync"
```

`run-local` executa `go run .` e repassa os argumentos (você pode passar direto, ex. `make run-local self sync`, ou usar `ARGS="..."`). O config e os plugins são os já instalados no seu usuário.

---

## 2. Compilar e executar o binário (também usa plugins instalados)

Gera o executável em `bin/mb` e roda com o mesmo config real (plugins instalados).

```bash
make build
./bin/mb --help
./bin/mb plugins list
./bin/mb <categoria> <comando>
```

Ou em um comando só:

```bash
make run plugins list
make run tools meu-plugin
# Ou: make run ARGS="self sync"
```

`run` faz `make build` e depois executa o binário com os argumentos (passados direto ou via `ARGS="..."`). Continua usando o config/plugins do usuário.

---

## Resumo

| Objetivo | Comando | Plugins usados |
|----------|---------|----------------|
| Rodar local com plugins instalados | `make run-local [args...]` ou `make run-local ARGS="..."` | Config real (~/.config/mb) |
| Build + rodar com plugins instalados | `make run [args...]` ou `make run ARGS="..."` | Config real |

Exemplos de argumentos: `self sync`, `plugins list`, `tools meu-plugin`. Para usar os plugins de exemplo do repositório sem copiá-los, use **`make install-examples`** (registra cada pasta de `examples/plugins` com `mb plugins add`); depois rode `make run self sync`.
