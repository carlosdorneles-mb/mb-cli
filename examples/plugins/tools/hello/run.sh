#!/bin/sh
# Plugin: tools/hello

# Log que respeita MB_QUIET e MB_VERBOSE.
# Uso: mb_log <level> <mensagem...>
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

log none "Olá, Mundo!"
log debug "Mensagem de debug (visível só com mb -v tools hello)"
log info "Oi essa é uma mensagem de informação indo para o gum log"
log warn "Aviso de exemplo"
log error "Exemplo de erro"
log fatal "Exemplo de fatal"
