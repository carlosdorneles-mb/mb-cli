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
make run-local ARGS="plugins list"
make run-local ARGS="self sync"
make run-local ARGS="tools meu-plugin"
```

`run-local` executa `go run .` e repassa `ARGS`; o config e os plugins são os já instalados no seu usuário.

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
make run ARGS="plugins list"
make run ARGS="tools meu-plugin"
```

`run` faz `make build` e depois `./bin/mb $(ARGS)`; continua usando o config/plugins do usuário.

---

## 3. Executar em sandbox (config isolado, sem ver plugins instalados)

Para **não** usar seu `~/.config/mb` (por exemplo em testes ou para não misturar com plugins reais), use um diretório temporário como config:

```bash
# Uma execução
XDG_CONFIG_HOME=/tmp/mb-test go run . self sync
XDG_CONFIG_HOME=/tmp/mb-test go run . plugins list
```

Várias execuções na mesma sessão:

```bash
export XDG_CONFIG_HOME=/tmp/mb-test
go run . self sync
go run . plugins list
go run . self env set MY_VAR value
```

**Via Makefile** (usa `/tmp/mb-sandbox` por padrão):

```bash
make run-sandbox ARGS="self sync"
make run-sandbox ARGS="plugins list"
```

No sandbox o CLI **não** enxerga os plugins já instalados; plugins e cache ficam só no diretório temporário.

---

## Resumo

| Objetivo | Comando | Plugins usados |
|----------|---------|----------------|
| Rodar local com plugins instalados | `go run . ARGS` ou `make run-local ARGS="..."` | Config real (~/.config/mb) |
| Build + rodar com plugins instalados | `make run ARGS="..."` | Config real |
| Rodar em sandbox (config isolado) | `make run-sandbox ARGS="..."` ou `XDG_CONFIG_HOME=/tmp/... go run . ARGS` | Só o dir temporário |

Substitua `ARGS` por qualquer comando do MB: `self sync`, `plugins list`, `--quiet self list`, `tools meu-plugin`, etc.
