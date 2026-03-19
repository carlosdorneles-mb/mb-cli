#!/bin/bash

# Snap Store helper: install, update, remove and query Snap applications.
# Compatible with Linux systems where snapd is available.
#
# Usage:
#   . "$MB_HELPERS_PATH/snap.sh"
#
# Public functions:
#   snap_is_available           - Check if Snap is installed
#   snap_is_installed           - Check if an app is installed
#   snap_get_installed_version  - Get installed version of an app
#   snap_get_latest_version     - Get latest available version
#   snap_install                - Install an application from Snap Store
#   snap_update                 - Update an installed application
#   snap_uninstall              - Remove an installed application
#   snap_refresh_metadata       - Refresh Snap Store metadata
#   snap_info                   - Get detailed information about a Snap package
#   snap_list_installed         - List all installed Snap packages

. "$MB_HELPERS_PATH/log.sh"

# Returns 0 if Snap is installed on the system, 1 otherwise.
# Usage:
#   snap_is_available
# Example:
#   if snap_is_available; then echo "snap is available"; fi
snap_is_available() {
    command -v snap &> /dev/null
}

# Refreshes Snap Store repository metadata to ensure latest version info is available.
# Returns 0 always (metadata update is not critical).
# Usage:
#   snap_refresh_metadata
# Example:
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

# Returns 0 if the given Snap application is installed, 1 otherwise.
# Usage:
#   snap_is_installed <app_name>
# Example:
#   if snap_is_installed "podman-desktop"; then echo "installed"; fi
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

# Prints the installed version of a Snap application, or "unknown" if not found.
# Usage:
#   snap_get_installed_version <app_name>
# Example:
#   version=$(snap_get_installed_version "podman-desktop")
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

# Prints the latest available version of an application from Snap Store, or "unknown" if not found.
# Returns 0 if version was found, 1 on error.
# Usage:
#   snap_get_latest_version <app_name>
# Example:
#   latest=$(snap_get_latest_version "podman-desktop")
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

# Installs an application from Snap Store. Returns 0 on success, 1 on error.
# Usage:
#   snap_install <app_name> [friendly_name] [channel] [classic]
# Example:
#   snap_install "podman-desktop" "Podman Desktop" "stable" "false"
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

# Updates an installed Snap application. Returns 0 if up to date or updated, 1 on error.
# Usage:
#   snap_update <app_name> [friendly_name] [channel]
# Example:
#   snap_update "podman-desktop" "Podman Desktop"
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
    log debug "Versão atual: $current_version"

    # Update the application
    log info "Atualizando $friendly_name via Snap..."

    local update_cmd="snap refresh $app_name --channel=$channel"
    if ! eval "sudo $update_cmd"; then
        log error "Falha ao atualizar $friendly_name via Snap"
        return 1
    fi

    # Check new version
    local new_version=$(snap_get_installed_version "$app_name")

    if [ "$current_version" = "$new_version" ]; then
        log info "$friendly_name já estava na versão mais recente ($new_version)"
    else
        log info "$friendly_name atualizado com sucesso para versão $new_version!"
    fi

    return 0
}

# Removes an installed Snap application. Returns 0 on success or if not installed, 1 on error.
# Usage:
#   snap_uninstall <app_name> [friendly_name]
# Example:
#   snap_uninstall "podman-desktop" "Podman Desktop"
snap_uninstall() {
    local app_name="${1:-}"
    local friendly_name="${2:-$app_name}"

    if [ -z "$app_name" ]; then
        log error "Nome da aplicação é obrigatório"
        return 1
    fi

    # Check if installed
    if ! snap_is_installed "$app_name"; then
        log debug "$friendly_name não está instalado"
        return 0
    fi

    log info "Removendo $friendly_name..."
    log debug "Nome da aplicação: $app_name"

    if ! sudo snap remove --purge "$app_name" 2> /dev/null; then
        log error "Falha ao remover $friendly_name via Snap"
        return 1
    fi

    # Verify removal
    if ! snap_is_installed "$app_name"; then
        log info "$friendly_name removido com sucesso!"
        return 0
    else
        log error "Falha ao remover $friendly_name completamente"
        return 1
    fi
}

# Prints detailed information about a Snap package from the store.
# Returns 0 on success, 1 on error.
# Usage:
#   snap_info <app_name>
# Example:
#   snap_info "podman-desktop"
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

# Lists all installed Snap packages with their versions.
# Returns 0 on success, 1 if Snap is not available.
# Usage:
#   snap_list_installed
# Example:
#   snap_list_installed
snap_list_installed() {
    if ! snap_is_available; then
        log error "Snap não está instalado"
        return 1
    fi

    snap list
}
