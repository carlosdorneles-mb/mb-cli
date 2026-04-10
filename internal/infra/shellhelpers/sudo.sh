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
#   is_root                      — teste silencioso (0 = pode usar sudo sem prompt ou já é root).
#   warn_and_skip_without_sudo   — se não is_root: log warn (PT-BR) e retorna 86 (skip sudo; ver contrato abaixo).
#   check_sudo                   — igual a is_root; se falhar, log warn e retorna 1.
#   required_sudo                — exige sudo interativo, ou modo --optional para seguir sem abortar.
#
# Contrato de códigos de saída para plugins shell (alinhado a mb-cli-plugins/tools/update-all.sh e
# constantes Go em mb/internal/shared/pluginexit):
#   86 — MB_EXIT_UPDATE_SKIPPED_SUDO: sem privilégio efetivo para gestores de pacote (apt/dnf/yum/pacman).
#        No batch --update-all, não conta como falha; o pai avisa para repetir com sudo.
#        Em mb tools <plugin> --install|--update|--uninstall direto, o utilizador vê o log warn deste helper
#        no stderr; o processo mb pode ainda mostrar ERRO com "exit status 86".
#   87 — MB_EXIT_UPDATE_SKIPPED_NOT_INSTALLED: ferramenta não instalada ao atualizar; ignorado no batch.
#   88 — MB_EXIT_INSTALL_ALREADY_INSTALLED: em lote (MB_TOOLS_INSTALL_BATCH), install.sh pode devolver 88
#        quando a ferramenta já está instalada; o install-many trata como sucesso sem nova instalação.
#        Em instalação única costuma-se exit 0 para não sinalizar erro.
# Fora do update-all.sh, usar sempre return "${MB_EXIT_UPDATE_SKIPPED_SUDO:-86}" (e :-87) para defaults.

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

# Se já há privilégio efetivo (root ou sudo -n), retorna 0. Caso contrário regista aviso formal em
# PT-BR no stderr e devolve MB_EXIT_UPDATE_SKIPPED_SUDO (86 por omissão), para alinhar com
# mb tools --update-all e permitir mensagem útil quando o utilizador corre mb tools … diretamente.
#
# Args opcional:
#   $1 — texto curto de contexto (ex.: nome da ferramenta) anteposto à mensagem padrão.
#
# Usage (no início de install_linux / update_linux / uninstall_linux):
#   warn_and_skip_without_sudo || return $?
#   warn_and_skip_without_sudo "Redis CLI" || return $?
#
# Returns:
#   0 — pode prosseguir com apt/dnf/etc.
#   86 — ou valor de MB_EXIT_UPDATE_SKIPPED_SUDO quando definido (fora do batch, usar :-86 no return).
warn_and_skip_without_sudo() {
    if is_root; then
        return 0
    fi
    local ctx="${1:-}"
    local base="Instalação, atualização ou remoção de pacotes do sistema requer privilégios de administrador (sudo)."
    local tail=" Execute o mesmo comando com sudo (por exemplo: sudo mb tools <ferramenta> --install ou --update), ou execute sudo -v antes de repetir."
    if [ -n "$ctx" ]; then
        log warn "$ctx — $base$tail" >&2
    else
        log warn "$base$tail" >&2
    fi
    return "${MB_EXIT_UPDATE_SKIPPED_SUDO:-86}"
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
#   1) Se is_root passa (root ou sudo -n), retorna 0 sem logs.
#   2) Caso contrário, executa `sudo -v` uma única vez (pode pedir senha).
#   3) Se `sudo -v` falhar:
#      - com --optional: registra warn (via check_sudo) e retorna 0.
#      - sem --optional: registra warn (via check_sudo), depois error e encerra com exit 1.
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

    if sudo -v; then
        return 0
    fi
    
    if [ "$optional" = true ]; then
        if [ -n "$cmd_context" ]; then
            check_sudo "Executando sem privilégios de superusuário (sudo): algumas funcionalidades de \"$cmd_context\" podem não funcionar ou ficar indisponíveis."
        else
            check_sudo "Executando sem privilégios de superusuário (sudo): algumas funcionalidades deste comando podem não funcionar ou ficar indisponíveis."
        fi

        return 0
    fi

    log error "Falha ao obter privilégios de sudo"
    exit 1
}
