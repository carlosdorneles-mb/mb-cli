---
sidebar_position: 4
---

# Flags globais

O MB CLI oferece algumas flags que valem para qualquer comando e afetam o nível de saída e o ambiente. **`-v` e `-q` podem ser usados em qualquer posição:** antes ou depois do subcomando (por exemplo `mb -v tools hello` ou `mb tools hello -v`). Em ambos os casos o plugin recebe `MB_VERBOSE`/`MB_QUIET` conforme a flag.

## --verbose / -v

**O que faz:** Ativa saída mais verbosa. O CLI pode exibir mensagens adicionais de diagnóstico ou logs que ajudam a entender o que está acontecendo.

**Quando usar:** Para depurar um problema, acompanhar o fluxo de um comando ou quando você quer mais detalhes na saída.

Ao executar um **comando de plugin**, o CLI define no ambiente do processo do plugin a variável **`MB_VERBOSE=1`**. O plugin pode usá-la para exibir mais logs (por exemplo, nível debug).

Exemplo:

```bash
mb -v plugins list
mb --verbose self sync
mb -v tools meu-comando
```

## --quiet / -q

**O que faz:** Reduz ou suprime mensagens informativas. O CLI evita imprimir avisos ou mensagens de progresso que não sejam estritamente necessárias.

**Quando usar:** Em scripts ou quando você quer apenas o resultado (por exemplo, saída de um plugin) sem mensagens extras do próprio CLI.

Ao executar um **comando de plugin**, o CLI define no ambiente do processo do plugin a variável **`MB_QUIET=1`**. O plugin pode usá-la para suprimir logs informativos e exibir apenas o essencial (ou apenas erros).

Exemplo:

```bash
mb -q plugins list
mb --quiet self sync
mb -q tools meu-comando
```

## Uso em plugins de shell

Em plugins escritos em shell, você pode ler `MB_VERBOSE` e `MB_QUIET` para decidir se imprime mensagens e em qual nível. Assim o plugin respeita a preferência do usuário ao usar `-v` ou `-q`.

- **`MB_QUIET=1`** — O usuário pediu saída mínima. Evite chamar `gum log` para mensagens informativas; só mostre erros se fizer sentido.
- **`MB_VERBOSE=1`** — O usuário pediu mais detalhes. Você pode incluir logs em nível debug ou mensagens de diagnóstico.

Exemplo: função `log` que centraliza a lógica e evita repetir condicionais em cada chamada. Uso: `log <level> <mensagem...>`. Níveis: `none`, `debug`, `info`, `warn`, `error`, `fatal`. Com `MB_QUIET=1` só exibe error e fatal; com `MB_VERBOSE=1` exibe todos (incluindo debug); caso contrário debug é omitido.

```sh
# Log que respeita MB_QUIET e MB_VERBOSE.
# Uso: log <level> <mensagem...>
# Níveis: none, debug, info, warn, error, fatal
# - MB_QUIET=1: só exibe error e fatal
# - MB_VERBOSE=1: exibe todos (incluindo debug); caso contrário debug é omitido
log() {
  level=$1
  shift
  [ -z "$*" ] && return 0

  if [ -n "$MB_QUIET" ]; then
    case "$level" in
      error|fatal) ;;
      *) return 0 ;;
    esac
  fi

  if [ -z "$MB_VERBOSE" ] && [ "$level" = "debug" ]; then
    return 0
  fi

  gum log -l "$level" "$@"
}
```

Depois basta chamar `log info "Processando..."`, `log debug "Detalhe: $var"`, etc., sem repetir `if` em cada linha.

## --env-file e --env / -e

- **`--env-file <path>`** — Define um arquivo de variáveis de ambiente (formato `.env`) que será carregado e mesclado ao ambiente antes de executar um plugin. Útil para manter configurações em um arquivo separado.
- **`--env KEY=VALUE`** — Injeta uma variável no processo do plugin. Pode ser repetido várias vezes. Tem a maior precedência em relação aos outros meios de definir variáveis.

Para a ordem completa de precedência e como usar defaults com `mb self env`, veja [Variáveis de ambiente](./variaveis-ambiente.md).
