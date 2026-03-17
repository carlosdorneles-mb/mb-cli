#!/bin/sh

# MB CLI simple memory helper (key/value), persisted in TMPDIR.
# Compatible with Linux and macOS (POSIX sh).
#
# Usage:
#   mem_set <namespace> <key> <value...>
#   mem_get <namespace> <key> [default]
#   mem_has <namespace> <key>
#   mem_unset <namespace> <key>
#   mem_clear_ns <namespace>

# Returns the base directory for memory storage: ${TMPDIR:-/tmp}/mb/memory.
mem_root() {
  _mem_root="${TMPDIR:-/tmp}/mb/memory"
  printf "%s" "$_mem_root"
}

# Returns 0 if the token only contains letters, digits, dot, underscore, or hyphen; 1 otherwise.
mem__is_valid_token() {
  case "$1" in
    ""|*[!A-Za-z0-9._-]*) return 1 ;;
    *) return 0 ;;
  esac
}

# Returns the file path for a given namespace and key. Returns 2 if either token is invalid.
mem__path() {
  _mem_ns=$1
  _mem_key=$2

  mem__is_valid_token "$_mem_ns" || return 2
  mem__is_valid_token "$_mem_key" || return 2

  printf "%s/%s/%s" "$(mem_root)" "$_mem_ns" "$_mem_key"
}

# Saves or overwrites a value for the given namespace and key.
# Usage:
#   mem_set <namespace> <key> <value...>
# Example:
#   mem_set tools.deploy cluster prod
mem_set() {
  _mem_ns=$1
  _mem_key=$2
  shift 2
  _mem_value=$*

  _mem_path=$(mem__path "$_mem_ns" "$_mem_key") || return $?
  _mem_dir=$(dirname "$_mem_path")
  _mem_tmp="${_mem_path}.tmp.$$"

  # Use a subshell to avoid changing the caller script global umask.
  (
    umask 077
    mkdir -p "$_mem_dir" &&
      printf "%s" "$_mem_value" > "$_mem_tmp" &&
      mv "$_mem_tmp" "$_mem_path"
  ) || return 1
}

# Reads the value for the given namespace and key. Returns default (or empty) if the key does not exist.
# Usage:
#   mem_get <namespace> <key> [default]
# Example:
#   cluster="$(mem_get tools.deploy cluster dev)"
mem_get() {
  _mem_ns=$1
  _mem_key=$2
  _mem_default=$3

  _mem_path=$(mem__path "$_mem_ns" "$_mem_key") || return $?
  if [ -f "$_mem_path" ]; then
    cat "$_mem_path"
  else
    printf "%s" "$_mem_default"
  fi
}

# Returns 0 if the key exists, 1 otherwise.
# Usage:
#   mem_has <namespace> <key>
# Example:
#   if mem_has tools.deploy cluster; then
#     echo "found"
#   fi
mem_has() {
  _mem_path=$(mem__path "$1" "$2") || return $?
  [ -f "$_mem_path" ]
}

# Removes a specific key from memory.
# Usage:
#   mem_unset <namespace> <key>
# Example:
#   mem_unset tools.deploy cluster
mem_unset() {
  _mem_path=$(mem__path "$1" "$2") || return $?
  rm -f "$_mem_path"
}

# Removes all keys under a namespace.
# Usage:
#   mem_clear_ns <namespace>
# Example:
#   mem_clear_ns tools.deploy
mem_clear_ns() {
  _mem_ns=$1
  mem__is_valid_token "$_mem_ns" || return 2
  rm -rf "$(mem_root)/$_mem_ns"
}