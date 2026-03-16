# Helper de log do MB CLI. Sourcear via . "$MB_HELPERS_PATH/log.sh" ou tudo via all.sh.
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
