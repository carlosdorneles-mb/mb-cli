#!/bin/bash

# Returns 0 if the effective user is root.
is_root() {
    [ "${EUID:-$(id -u)}" -eq 0 ]
}

# Check if the script is running with superuser privileges
# Usage:
#   check_sudo
#   check_sudo "mensagem personalizada para o warning"
#
# Args:
#   texto opcional — mensagem exibida em log warn quando não for root; se omitido, usa a mensagem padrão.
#
# Returns:
#   0 - If running as root
#   1 - If not running with root privileges
check_sudo() {
    local default_msg="Este comando requer privilégios de superusuário (sudo). Autentique-se quando solicitado ou execute com sudo."
    if ! is_root; then
        log warn "${1:-$default_msg}" >&2
        return 1
    fi
    return 0
}

# Ensure the script has superuser privileges (refresh sudo credentials when needed).
# Com --optional: tenta sudo -v (pede a senha); se o usuário não autenticar ou falhar,
# emite o aviso de execução sem sudo e segue o script.
#
# Usage:
#   required_sudo
#   required_sudo --optional
#   required_sudo --optional "descrição do comando ou contexto"
#
# Args:
#   --optional       Tenta elevar com sudo; se não der, avisa e continua sem falhar.
#   texto opcional   Após --optional, descreve o comando para contextualizar o aviso.
required_sudo() {
    local optional=false
    local cmd_context=""

    if [ "${1:-}" = "--optional" ]; then
        optional=true
        shift
        if [ $# -gt 0 ]; then
            cmd_context=$1
            shift
        fi
    fi

    if check_sudo; then
        return 0
    fi

    if [ "$optional" = true ]; then
        if sudo -v; then
            return 0
        fi
        if [ -n "$cmd_context" ]; then
            log warn "Executando sem privilégios de superusuário (sudo): algumas funcionalidades de \"$cmd_context\" podem não funcionar ou ficar indisponíveis." >&2
        else
            log warn "Executando sem privilégios de superusuário (sudo): algumas funcionalidades deste comando podem não funcionar ou ficar indisponíveis." >&2
        fi
        return 0
    fi

    check_sudo || true
    sudo -v || {
        log error "Falha ao obter privilégios de sudo"
        exit 1
    }
}
