---
sidebar_position: 7
---

# Helpers de shell

O MB CLI injeta no ambiente dos plugins a variável **`MB_HELPERS_PATH`**, que aponta para o **diretório** dos helpers de shell (`~/.config/mb/lib/shell`).

## Como carregar

No início do script do plugin (por exemplo em `run.sh`), importe o que precisar:

- **Todos os helpers:** `. "$MB_HELPERS_PATH/all.sh"`
- **Só o helper de log:** `. "$MB_HELPERS_PATH/log.sh"`
- **Só o helper de memória:** `. "$MB_HELPERS_PATH/memory.sh"`
- **Só o helper de string:** `. "$MB_HELPERS_PATH/string.sh"`
- **Só o helper de Kubernetes:** `. "$MB_HELPERS_PATH/kubernetes.sh"`

Exemplo:

```sh
#!/bin/sh
. "$MB_HELPERS_PATH/all.sh"

# A partir daqui você pode usar os helpers
log info "Olá!"
```

O diretório e os arquivos são criados ou atualizados quando você executa **`mb self sync`** (ou ao adicionar/atualizar plugins, que disparam o sync). Se os helpers ainda não existirem, execute `mb self sync` antes de usá-los nos seus plugins. Ao atualizar o CLI para uma versão que altere os helpers, o próximo `mb self sync` atualiza os arquivos em `lib/shell` automaticamente (o CLI compara um checksum do conteúdo embutido com o arquivo `.checksum` nesse diretório).

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

### memory

Helper de memória simples (chave/valor) para scripts de plugin.

Ele salva dados em arquivos no diretório temporário do sistema (`${TMPDIR:-/tmp}/mb/memory`) usando a estrutura `namespace/key`. Isso permite reaproveitar respostas curtas do usuário em execuções futuras do mesmo plugin.

Como funciona:

- Cada valor fica em um arquivo: `${TMPDIR:-/tmp}/mb/memory/<namespace>/<key>`.
- O valor é sobrescrito quando você chama `mem_set` novamente para a mesma chave.
- A escrita é feita com arquivo temporário + `mv` (atômica) para reduzir risco de arquivo parcial.
- `namespace` e `key` aceitam somente letras, números, `.`, `_` e `-`.
- Por estar em `tmp`, o conteúdo pode ser removido pelo sistema (reboot/limpeza automática).

**Uso:**

- `mem_set <namespace> <key> <valor...>`
- `mem_get <namespace> <key> [default]`
- `mem_has <namespace> <key>`
- `mem_unset <namespace> <key>`
- `mem_clear_ns <namespace>`

**Comandos disponíveis:**

- `mem_set`: cria ou atualiza um valor.
	Ex.: `mem_set tools.deploy cluster prod`
- `mem_get`: lê um valor; se não existir, retorna o `default` (ou vazio).
	Ex.: `cluster="$(mem_get tools.deploy cluster dev)"`
- `mem_has`: verifica se a chave existe (ideal para `if`).
	Ex.: `if mem_has tools.deploy cluster; then ... fi`
- `mem_unset`: remove uma chave específica.
	Ex.: `mem_unset tools.deploy cluster`
- `mem_clear_ns`: remove todas as chaves de um namespace.
	Ex.: `mem_clear_ns tools.deploy`

Retornos:

- `0`: sucesso.
- `1`: ausência de chave em `mem_has` ou falha de I/O.
- `2`: `namespace`/`key` inválidos.

Exemplo:

```sh
. "$MB_HELPERS_PATH/memory.sh"

if ! mem_has "tools.deploy" "cluster"; then
	cluster="$(gum input --placeholder "Nome do cluster")"
	mem_set "tools.deploy" "cluster" "$cluster"
fi

cluster="$(mem_get "tools.deploy" "cluster" "dev")"
log info "Usando cluster: $cluster"
```

Observações:

- Esses dados ficam em `tmp` e podem ser removidos pelo sistema (por exemplo, em reboot ou limpeza automática).
- `namespace` e `key` aceitam somente letras, números, `.`, `_` e `-`.

### string

Helper de utilitários para manipulação de texto em scripts shell. Cobre substituição, conversão de case, trim, testes de conteúdo, manipulação de arrays CSV e conversão de booleano.

**Funções disponíveis:**

- `str_replace <input> <search> <replace>` — substitui todas as ocorrências de `search` por `replace` em `input` e imprime o resultado.
- `str_to_upper <texto>` — imprime o texto convertido para maiúsculas.
- `str_to_lower <texto>` — imprime o texto convertido para minúsculas.
- `str_trim <texto>` — imprime o texto sem espaços no início e no fim.
- `str_contains <texto> <substring>` — retorna `0` se `texto` contém `substring`, `1` caso contrário.
- `str_starts_with <texto> <prefixo>` — retorna `0` se `texto` começa com `prefixo`, `1` caso contrário.
- `str_parse_comma_separated <nome_array>` — percorre o array referenciado e divide cada elemento que contenha vírgula em elementos separados (modifica o array in-place).
- `str_join_to_comma_separated <nome_array>` — junta todos os elementos do array em um único elemento separado por vírgula (modifica o array in-place).
- `str_to_bool <valor>` — retorna `0` para valores verdadeiros (`true`, `1`, `on`, `yes`) e `1` para os demais.

Exemplo:

```sh
. "$MB_HELPERS_PATH/string.sh"

# Substituição e conversão
tag=$(str_to_lower "$(str_trim "  My-App  ")")
log info "Tag: $tag"  # my-app

# Testes condicionais
if str_starts_with "$tag" "my"; then
  log info "Tag começa com 'my'"
fi

# Booleano a partir de variável de ambiente
if str_to_bool "${DRY_RUN:-false}"; then
  log warn "Dry-run ativo, nenhuma alteração será feita"
fi
```

### kubernetes

Helper para operações básicas com `kubectl`: verificar se está instalado, checar existência de namespace e inspecionar o contexto ativo. Carrega `log.sh` automaticamente ao ser importado.

> **Requisito:** `kubectl` precisa estar instalado e configurado no `PATH`. Caso contrário, as funções logam um erro e, se `exit_on_error` for passado, encerram o script com `exit 1`.

**Funções disponíveis:**

- `kb_check_installed [exit_on_error]` — verifica se `kubectl` está disponível no `PATH`. Retorna `0` se encontrado, `1` se não. Com `exit_on_error`, encerra o script se não estiver instalado.
- `kb_check_namespace_exists <namespace> [exit_on_error]` — verifica se o namespace existe no cluster do contexto atual. Retorna `0` se existir, `1` se não. Com `exit_on_error`, encerra o script se não existir.
- `kb_get_current_context` — imprime o nome do contexto kubectl ativo (`kubectl config current-context`).
- `kb_print_current_context` — imprime o contexto atual no console com uma mensagem legível.

Exemplo:

```sh
. "$MB_HELPERS_PATH/kubernetes.sh"

# Garante que kubectl existe e que o namespace alvo também
kb_check_installed "exit_on_error"
kb_check_namespace_exists "production" "exit_on_error"

# Informa o contexto em uso antes de aplicar mudanças
kb_print_current_context
kubectl apply -f manifests/
```
