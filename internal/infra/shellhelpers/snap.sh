#!/bin/bash

# Helper Snap Store: instalar, atualizar, remover e consultar pacotes snap.
# Compatível com Linux onde o snapd está instalado e no PATH.
#
# Carrega log.sh (MB_HELPERS_PATH). Mensagens respeitam MB_QUIET e MB_VERBOSE.
#
# Uso:
#   . "$MB_HELPERS_PATH/snap.sh"
#
# Requisitos:
#   - Comando `snap` disponível para leituras; instalação, refresh e remove usam `sudo`
#     (o usuário precisa poder elevar privilégio quando solicitado).
#
# Funções públicas:
#   snap_is_available           — snap no sistema (exit 0/1)
#   snap_refresh_metadata       — atualiza lista de atualizações (snap refresh --list); não falha o script
#   snap_is_installed <nome>    — pacote instalado (grep em snap list)
#   snap_get_installed_version  — imprime versão instalada ou unknown
#   snap_get_latest_version     — lê trilha latest/stable de `snap info`
#   snap_install                — snap install (canal, classic opcional)
#   snap_update                 — snap refresh no canal indicado
#   snap_uninstall              — snap remove --purge
#   snap_info                   — saída bruta de snap info
#   snap_list_installed         — snap list

. "$MB_HELPERS_PATH/log.sh"

# Retorna 0 se o executável `snap` existe no PATH; caso contrário, 1.
#
# Usage:
#   snap_is_available
snap_is_available() {
    command -v snap &> /dev/null
}

# Atualiza a lista de revisões disponíveis (`snap refresh --list`). Erros são ignorados
# (log debug). Retorna sempre 0. Se snap não existir, não faz nada útil e retorna 0.
#
# Usage:
#   snap_refresh_metadata
snap_refresh_metadata() {
    if ! snap_is_available; then
        log debug "Snap não disponível, pulando atualização de metadados"
        return 0
    fi

    log debug "Atualizando metadados do Snap Store..."
    snap refresh --list &> /dev/null || log debug "Metadados já estão atualizados"
    return 0
}

# Retorna 0 se o snap <app_name> aparece em `snap list` (nome na primeira coluna).
#
# Usage:
#   snap_is_installed <app_name>
#
# Returns:
#   1 — nome vazio, snap indisponível ou pacote ausente.
snap_is_installed() {
    local app_name="${1:-}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        return 1
    fi

    if ! snap_is_available; then
        return 1
    fi

    snap list 2> /dev/null | grep -q "^${app_name}\s"
}

# Imprime a revisão/versão na segunda coluna de `snap list <app>` ou "unknown".
# Não falha o shell: nome vazio ou não instalado → stdout "unknown", exit 0.
#
# Usage:
#   version=$(snap_get_installed_version <app_name>)
snap_get_installed_version() {
    local app_name="${1:-}"

    if [ -z "$app_name" ]; then
        echo "unknown"
        return 0
    fi

    if ! snap_is_installed "$app_name"; then
        echo "unknown"
        return 0
    fi

    local version=$(snap list "$app_name" 2> /dev/null | tail -n +2 | awk '{print $2}' || echo "unknown")
    echo "$version"
}

# Obtém a versão publicada em `snap info` na linha `latest/stable:` (segundo campo).
# Em stdout: versão ou "unknown". Exit 1 se nome vazio, snap ausente ou linha não encontrada.
#
# Usage:
#   latest=$(snap_get_latest_version <app_name>)
snap_get_latest_version() {
    local app_name="${1:-}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        echo "unknown"
        return 1
    fi

    if ! snap_is_available; then
        log error "Snap não está instalado"
        echo "unknown"
        return 1
    fi

    # Try to get from snap info
    local version=$(snap info "$app_name" 2> /dev/null | grep -E "^\s*latest/stable:" | awk '{print $2}' || echo "")

    # If not found, return unknown
    if [ -z "$version" ]; then
        echo "unknown"
        return 1
    fi

    echo "$version"
    return 0
}

# Instala um snap. Se já instalado, log info e retorna 0. Usa `sudo snap install`.
#
# Usage:
#   snap_install <app_name> [friendly_name] [channel] [classic]
#
# Args:
#   app_name      — nome do pacote no Snap Store (obrigatório).
#   friendly_name — rótulo nos logs (padrão: app_name).
#   channel       — padrão stable (passado a --channel=).
#   classic       — string "true" adiciona --classic; qualquer outro valor omite.
#
# Returns:
#   0 — já instalado ou instalação verificada com sucesso.
#   1 — erro (snap ausente, falha do install ou pacote não aparece após install).
snap_install() {
    local app_name="${1:-}"
    local friendly_name="${2:-$app_name}"
    local channel="${3:-stable}"
    local classic="${4:-false}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        return 1
    fi

    # Check if Snap is available
    if ! snap_is_available; then
        log error "Snap não está instalado. Por favor, instale o Snap primeiro."
        log info "Veja: https://snapcraft.io/docs/installing-snapd"
        return 1
    fi

    # Check if already installed
    if snap_is_installed "$app_name"; then
        local version=$(snap_get_installed_version "$app_name")
        log info "$friendly_name $version já está instalado"
        return 0
    fi

    # Update metadata
    snap_refresh_metadata

    # Install the application
    log info "Instalando $friendly_name via Snap..."
    log debug "Nome da aplicação: $app_name"
    log debug "Canal: $channel"

    local install_cmd="snap install $app_name --channel=$channel"
    if [ "$classic" = "true" ]; then
        install_cmd="$install_cmd --classic"
        log debug "Modo: classic"
    fi

    if ! eval "sudo $install_cmd"; then
        log error "Falha ao instalar $friendly_name via Snap"
        return 1
    fi

    # Verify installation
    if snap_is_installed "$app_name"; then
        local version=$(snap_get_installed_version "$app_name")
        return 0
    else
        log error "$friendly_name foi instalado mas não está disponível"
        return 1
    fi
}

# Atualiza um snap instalado com `sudo snap refresh`. Se versão instalada já coincide com
# a latest/stable obtida por snap_get_latest_version, só loga e retorna 0.
#
# Usage:
#   snap_update <app_name> [friendly_name] [channel]
#
# Args:
#   channel — repassado a snap refresh (padrão stable).
#
# Returns:
#   0 — já estava na versão mais recente ou refresh concluído com sucesso.
#   1 — não instalado ou falha no refresh.
snap_update() {
    local app_name="${1:-}"
    local friendly_name="${2:-$app_name}"
    local channel="${3:-stable}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        return 1
    fi

    # Check if installed
    if ! snap_is_installed "$app_name"; then
        log error "$friendly_name não está instalado"
        return 1
    fi

    local current_version=$(snap_get_installed_version "$app_name")
    local latest_version=$(snap_get_latest_version "$app_name")

    log debug "Versão atual é $current_version, a mais recente é $latest_version"

    if [ "$current_version" = "$latest_version" ]; then
        log info "$friendly_name já estava na versão mais recente ($latest_version)"
    else
        log info "Atualizando $friendly_name via Snap..."

        if ! eval "sudo snap refresh $app_name --channel=$channel"; then
            log error "Falha ao atualizar $friendly_name via Snap"
            return 1
        fi

        log info "$friendly_name atualizado com sucesso para versão $latest_version!"
    fi

    return 0
}

# Remove um snap instalado com `sudo snap remove --purge`. Se não estiver instalado, retorna 0
# (log debug).
#
# Usage:
#   snap_uninstall <app_name> [friendly_name]
#
# Returns:
#   1 — nome vazio ou falha do snap remove.
snap_uninstall() {
    local app_name="${1:-}"
    local friendly_name="${2:-$app_name}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        return 1
    fi

    if ! snap_is_installed "$app_name"; then
        log debug "$friendly_name não está instalado"
        return 0
    fi

    log info "Removendo $friendly_name..."

    if ! sudo snap remove --purge "$app_name" 2> /dev/null; then
        log error "Falha ao remover $friendly_name via Snap"
        return 1
    fi

    log info "$friendly_name removido com sucesso!"
    return 0
}

# Executa `snap info <app_name>` na stdout (sem filtrar). Falha se snap indisponível ou nome vazio.
#
# Usage:
#   snap_info <app_name>
snap_info() {
    local app_name="${1:-}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        return 1
    fi

    if ! snap_is_available; then
        log error "Snap não está instalado"
        return 1
    fi

    snap info "$app_name"
}

# Executa `snap list` na stdout. Retorna 1 se o comando snap não existir.
#
# Usage:
#   snap_list_installed
snap_list_installed() {
    if ! snap_is_available; then
        log error "Snap não está instalado"
        return 1
    fi

    snap list
}
