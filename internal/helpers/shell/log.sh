#!/bin/sh

# Logs a message at the given level, respecting MB_QUIET and MB_VERBOSE.
# Source via . "$MB_HELPERS_PATH/log.sh" or load everything via all.sh.
# Levels: none, debug, info, warn, error, fatal.
# - MB_QUIET=1: shows only error and fatal.
# - MB_VERBOSE=1: shows all levels (including debug); otherwise debug is omitted.
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

  gum log -l "$level" "$@"
}
