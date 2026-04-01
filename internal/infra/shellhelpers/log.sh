#!/bin/bash

# Logs a message at the given level, respecting MB_QUIET and MB_VERBOSE.
# Source via . "$MB_HELPERS_PATH/log.sh" or load everything via all.sh.
# Levels: none, debug, info, warn, error, fatal, output, print.
#   - log output … / log print … — mensagem ao gum com nível de apresentação none (texto “limpo”).
# - MB_QUIET=1: shows only error and fatal.
# - MB_VERBOSE=1: shows all levels (including debug); otherwise debug is omitted.
# - MB_LOG_OUTPUT: quando definida (qualquer valor não vazio), o nível passado ao gum log
#   é none para qualquer chamada (como output/print). MB_QUIET / MB_VERBOSE continuam a usar
#   o nível declarado na chamada para filtrar antes de chegar ao gum.
# Usage:
#   log <level> <message...>
# Example:
#   log info "Deployment started"
#   log error "Something went wrong"
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

  gum_level=$level
  if [ "$level" = "output" ] || [ "$level" = "print" ] || [ -n "$MB_LOG_OUTPUT" ]; then
    gum_level=none
  fi

  gum log -sl "$gum_level" "$@"
}
