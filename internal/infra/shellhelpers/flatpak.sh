#!/bin/bash

# Flatpak helper: install, update, remove and query applications from Flathub.
# Compatible with Linux systems where Flatpak is available.
#
# Usage:
#   . "$MB_HELPERS_PATH/flatpak.sh"
#
# Public functions:
#   flatpak_is_available           - Check if Flatpak is installed
#   flatpak_ensure_flathub         - Ensure Flathub is configured
#   flatpak_is_installed           - Check if an app is installed
#   flatpak_get_installed_version  - Get installed version of an app
#   flatpak_get_latest_version     - Get latest available version
#   flatpak_install                - Install an application from Flathub
#   flatpak_update                 - Update an installed application
#   flatpak_uninstall              - Remove an installed application
#   flatpak_update_metadata        - Update Flathub metadata

. "$MB_HELPERS_PATH/log.sh"

# Returns 0 if Flatpak is installed on the system, 1 otherwise.
# Usage:
#   flatpak_is_available
# Example:
#   if flatpak_is_available; then echo "flatpak found"; fi
flatpak_is_available() {
    command -v flatpak &> /dev/null
}

# Ensures the Flathub repository is configured, adding it if needed (--user level).
# Returns 0 if already configured or successfully added, 1 on error.
# Usage:
#   flatpak_ensure_flathub
# Example:
#   flatpak_ensure_flathub
flatpak_ensure_flathub() {
    if ! flatpak_is_available; then
        log error "Flatpak não está instalado. Por favor, instale o Flatpak primeiro."
        log info "Veja: https://flatpak.org/setup/"
        return 1
    fi

    # Check if flathub is already added
    if flatpak remotes --user 2> /dev/null | grep -q "^flathub"; then
        log debug "Repositório Flathub já está configurado"
        return 0
    fi

    log info "Adicionando repositório Flathub..."
    if ! flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo; then
        log error "Falha ao adicionar repositório Flathub"
        return 1
    fi

    log info "Repositório Flathub adicionado com sucesso"
    return 0
}

# Updates Flathub repository metadata to ensure latest version info is available.
# Returns 0 always (metadata update is not critical).
# Usage:
#   flatpak_update_metadata
# Example:
#   flatpak_update_metadata
flatpak_update_metadata() {
    if ! flatpak_is_available; then
        log debug "Flatpak não disponível, pulando atualização de metadados"
        return 0
    fi

    log debug "Atualizando metadados do Flathub..."
    flatpak update --appstream --user 2> /dev/null || log debug "Metadados já estão atualizados"
    return 0
}

# Returns 0 if the given Flatpak application is installed, 1 otherwise.
# Usage:
#   flatpak_is_installed <app_id>
# Example:
#   if flatpak_is_installed "io.podman_desktop.PodmanDesktop"; then echo "installed"; fi
flatpak_is_installed() {
    local app_id="${1:-}"

    if [ -z "$app_id" ]; then
        log error "ID da aplicação é obrigatório"
        return 1
    fi

    if ! flatpak_is_available; then
        return 1
    fi

    flatpak list --user --app --columns=application 2> /dev/null | grep -q "^${app_id}$"
}

# Prints the installed version of a Flatpak application, or "unknown" if not found.
# Usage:
#   flatpak_get_installed_version <app_id>
# Example:
#   version=$(flatpak_get_installed_version "io.podman_desktop.PodmanDesktop")
flatpak_get_installed_version() {
    local app_id="${1:-}"

    if [ -z "$app_id" ]; then
        echo "unknown"
        return 0
    fi

    if ! flatpak_is_installed "$app_id"; then
        echo "unknown"
        return 0
    fi

    local version=$(flatpak info --user "$app_id" 2> /dev/null | grep -E "^\s*Version:" | head -1 | awk '{print $2}' || echo "unknown")
    echo "$version"
}

# Prints the latest available version of a Flatpak application from Flathub, or "unknown" if not found.
# Returns 0 if version was found, 1 on error.
# Usage:
#   flatpak_get_latest_version <app_id>
# Example:
#   latest=$(flatpak_get_latest_version "io.podman_desktop.PodmanDesktop")
flatpak_get_latest_version() {
    local app_id="${1:-}"

    if [ -z "$app_id" ]; then
        log error "ID da aplicação é obrigatório"
        echo "unknown"
        return 1
    fi

    if ! flatpak_is_available; then
        log error "Flatpak não está instalado"
        echo "unknown"
        return 1
    fi

    # Check if Flathub is configured
    if ! flatpak remotes --user 2> /dev/null | grep -q "^flathub"; then
        log error "Repositório Flathub não configurado"
        echo "unknown"
        return 1
    fi

    # Try to get from pending updates first
    local version=$(flatpak remote-ls --updates --user flathub --columns=application,version 2> /dev/null | grep "^${app_id}" | awk '{print $2}')

    # If not found in updates, search in remote-info
    if [ -z "$version" ]; then
        version=$(flatpak remote-info flathub --user "$app_id" 2> /dev/null | grep -E "^\s*Version:" | head -1 | awk '{print $2}')
    fi

    # If still not found, try Flathub API
    if [ -z "$version" ]; then
        log debug "Tentando obter versão via API do Flathub para $app_id"
        version=$(curl -fsSL "https://flathub.org/api/v2/appstream/${app_id}" 2> /dev/null | jq -r '.releases[0].version // empty' 2> /dev/null)
    fi

    # If still not found, return unknown
    if [ -z "$version" ]; then
        echo "unknown"
        return 1
    fi

    echo "$version"
    return 0
}

# Installs an application from Flathub. Returns 0 on success, 1 on error.
# Usage:
#   flatpak_install <app_id> [friendly_name]
# Example:
#   flatpak_install "io.podman_desktop.PodmanDesktop" "Podman Desktop"
flatpak_install() {
    local app_id="${1:-}"
    local app_name="${2:-$app_id}"

    if [ -z "$app_id" ]; then
        log error "Application ID is required"
        return 1
    fi

    # Ensure Flathub is configured
    if ! flatpak_ensure_flathub; then
        return 1
    fi

    # Check if already installed
    if flatpak_is_installed "$app_id"; then
        local version=$(flatpak_get_installed_version "$app_id")
        log info "$app_name $version já está instalado"
        return 0
    fi

    # Update metadata
    flatpak_update_metadata

    # Install the application
    log info "Instalando $app_name via Flatpak..."
    log debug "ID da aplicação: $app_id"

    if ! flatpak install -y --user flathub "$app_id"; then
        log error "Falha ao instalar $app_name via Flatpak"
        return 1
    fi

    # Verify installation
    if flatpak_is_installed "$app_id"; then
        local version=$(flatpak_get_installed_version "$app_id")
        return 0
    else
        log error "$app_name foi instalado mas não está disponível"
        return 1
    fi
}

# Updates an installed Flatpak application. Returns 0 if up to date or updated, 1 on error.
# Usage:
#   flatpak_update <app_id> [friendly_name]
# Example:
#   flatpak_update "io.podman_desktop.PodmanDesktop" "Podman Desktop"
flatpak_update() {
    local app_id="${1:-}"
    local app_name="${2:-$app_id}"

    if [ -z "$app_id" ]; then
        log error "ID da aplicação é obrigatório"
        return 1
    fi

    # Check if installed
    if ! flatpak_is_installed "$app_id"; then
        log error "$app_name não está instalado"
        return 1
    fi

    local current_version=$(flatpak_get_installed_version "$app_id")
    log debug "Versão atual: $current_version"

    # Update the application
    log info "Atualizando $app_name via Flatpak..."

    if ! flatpak update -y --user "$app_id"; then
        log error "Falha ao atualizar $app_name via Flatpak"
        return 1
    fi

    # Check new version
    local new_version=$(flatpak_get_installed_version "$app_id")

    if [ "$current_version" = "$new_version" ]; then
        log info "$app_name já estava na versão mais recente ($new_version)"
    else
        log info "$app_name atualizado com sucesso para versão $new_version!"
    fi

    return 0
}

# Removes an installed Flatpak application. Returns 0 on success or if not installed, 1 on error.
# Usage:
#   flatpak_uninstall <app_id> [friendly_name]
# Example:
#   flatpak_uninstall "io.podman_desktop.PodmanDesktop" "Podman Desktop"
flatpak_uninstall() {
    local app_id="${1:-}"
    local app_name="${2:-$app_id}"

    if [ -z "$app_id" ]; then
        log error "ID da aplicação é obrigatório"
        return 1
    fi

    # Check if installed
    if ! flatpak_is_installed "$app_id"; then
        log debug "$app_name não está instalado"
        return 0
    fi

    log info "Removendo $app_name..."
    log debug "ID da aplicação: $app_id"

    # Encerra processos da aplicação antes de desinstalar
    log debug "Encerrando processos de $app_name..."
    flatpak kill --user "$app_id" 2> /dev/null || true

    if ! flatpak uninstall --delete-data -y --user "$app_id" 2> /dev/null; then
        log error "Falha ao remover $app_name via Flatpak"
        return 1
    fi

    # Verify removal
    if ! flatpak_is_installed "$app_id"; then
        log info "$app_name removido com sucesso!"
        return 0
    else
        log error "Falha ao remover $app_name completamente"
        return 1
    fi
}
