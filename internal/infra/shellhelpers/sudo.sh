#!/bin/bash

# Helper de elevação de privilégio para scripts shell (root / sudo).
#
# Carrega log.sh (exige MB_HELPERS_PATH). Uso típico:
#   . "$MB_HELPERS_PATH/sudo.sh"
# Ou via all.sh (log.sh já foi carregado antes; source duplo de log.sh é inofensivo).
#
# Conceito: "tem privilégio efetivo" = processo como root OU sudo não interativo disponível
# (`sudo -n true`), ou seja, credencial em cache ou regra NOPASSWD — sem pedir senha no TTY.
#
# Funções públicas:
#   is_root       — teste silencioso (0 = pode usar sudo sem prompt ou já é root).
#   check_sudo    — igual a is_root; se falhar, log warn e retorna 1.
#   required_sudo — exige sudo interativo, ou modo --optional para seguir sem abortar.

. "$MB_HELPERS_PATH/log.sh"

# Retorna 0 se o processo já roda como root (EUID 0) ou se `sudo -n` aceita um comando
# sem prompt (credencial em cache / NOPASSWD). Retorna não zero caso contrário.
# Não escreve logs. Não solicita senha.
#
# Usage:
#   if is_root; then ...; fi
is_root() {
    [ "${EUID:-$(id -u)}" -eq 0 ] || sudo -n true 2>/dev/null
}

# Verifica privilégio efetivo (mesmo critério que is_root). Em caso de falha, registra
# log warn em stderr e retorna 1.
#
# Usage:
#   check_sudo
#   check_sudo "mensagem personalizada para o warning"
#
# Args:
#   texto opcional — mensagem do warn quando não há root/sudo -n; se omitido, mensagem padrão.
#
# Returns:
#   0 — root ou sudo não interativo disponível.
#   1 — caso contrário.
check_sudo() {
    local default_msg="Este comando requer privilégios de superusuário (sudo). Autentique-se quando solicitado ou execute com sudo."
    if ! is_root; then
        log warn "${1:-$default_msg}" >&2
        return 1
    fi
    return 0
}

# Garante credencial sudo para o restante do script, ou encerra (modo padrão), ou avisa e
# segue (modo --optional).
#
# Fluxo:
#   1) Se check_sudo passa (root ou sudo -n), retorna 0.
#   2) Com --optional: executa `sudo -v` (pode pedir senha). Falhou → warn e retorna 0.
#   3) Sem --optional: repete aviso via check_sudo, depois `sudo -v`; falhou → log error e exit 1.
#
# Usage:
#   required_sudo
#   required_sudo --optional
#   required_sudo --optional "descrição do comando ou contexto"
#
# Args:
#   --optional      — não encerra o script se o usuário não obtiver sudo.
#   texto (após --optional) — entra na mensagem de aviso quando segue sem sudo.
#
# Returns:
#   0 — em todos os caminhos que não chamam exit.
#
# Exit:
#   1 — apenas no modo obrigatório, se `sudo -v` falhar.
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
