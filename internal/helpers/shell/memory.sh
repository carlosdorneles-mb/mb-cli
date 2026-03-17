# Helper de memória simples (chave/valor) do MB CLI, persistido em TMPDIR.
# Compatível com Linux e macOS (POSIX sh).
#
# Uso:
#   mem_set <namespace> <key> <valor...>
#   mem_get <namespace> <key> [default]
#   mem_has <namespace> <key>
#   mem_unset <namespace> <key>
#   mem_clear_ns <namespace>

mem_root() {
  _mem_root="${TMPDIR:-/tmp}/mb/memory"
  printf "%s" "$_mem_root"
}

mem__is_valid_token() {
  case "$1" in
    ""|*[!A-Za-z0-9._-]*) return 1 ;;
    *) return 0 ;;
  esac
}

mem__path() {
  _mem_ns=$1
  _mem_key=$2

  mem__is_valid_token "$_mem_ns" || return 2
  mem__is_valid_token "$_mem_key" || return 2

  printf "%s/%s/%s" "$(mem_root)" "$_mem_ns" "$_mem_key"
}

mem_set() {
  _mem_ns=$1
  _mem_key=$2
  shift 2
  _mem_value=$*

  _mem_path=$(mem__path "$_mem_ns" "$_mem_key") || return $?
  _mem_dir=$(dirname "$_mem_path")
  _mem_tmp="${_mem_path}.tmp.$$"

  # Subshell para não alterar o umask global do script chamador.
  (
    umask 077
    mkdir -p "$_mem_dir" &&
      printf "%s" "$_mem_value" > "$_mem_tmp" &&
      mv "$_mem_tmp" "$_mem_path"
  ) || return 1
}

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

mem_has() {
  _mem_path=$(mem__path "$1" "$2") || return $?
  [ -f "$_mem_path" ]
}

mem_unset() {
  _mem_path=$(mem__path "$1" "$2") || return $?
  rm -f "$_mem_path"
}

mem_clear_ns() {
  _mem_ns=$1
  mem__is_valid_token "$_mem_ns" || return 2
  rm -rf "$(mem_root)/$_mem_ns"
}