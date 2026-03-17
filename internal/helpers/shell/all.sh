# Helpers de shell do MB CLI. Carrega todos os helpers.
# Ex.: . "$MB_HELPERS_PATH/all.sh"
# Usa MB_HELPERS_PATH (e não dirname "$0") porque, ao ser sourceado, $0 é o script do plugin.

. "${MB_HELPERS_PATH}/log.sh"
. "${MB_HELPERS_PATH}/memory.sh"
