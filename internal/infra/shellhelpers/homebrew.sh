#!/bin/bash

# Homebrew helper: install, update, remove and query casks and formulae on macOS.
# Compatible with macOS systems where Homebrew is available.
#
# Usage:
#   . "$MB_HELPERS_PATH/homebrew.sh"
#
# Public functions:
#   homebrew_is_available                  - Check if Homebrew is installed
#   homebrew_is_installed                  - Check if a cask is installed
#   homebrew_get_installed_version         - Get installed version of a cask
#   homebrew_get_latest_version            - Get latest available version of a cask
#   homebrew_install                       - Install a cask from Homebrew
#   homebrew_update                        - Update an installed cask
#   homebrew_uninstall                     - Remove an installed cask
#   homebrew_update_metadata               - Update Homebrew formulae metadata
#   homebrew_is_installed_formula          - Check if a formula is installed
#   homebrew_get_installed_version_formula - Get installed version of a formula
#   homebrew_get_latest_version_formula    - Get latest available version of a formula
#   homebrew_install_formula               - Install a formula from Homebrew
#   homebrew_update_formula                - Update an installed formula
#   homebrew_uninstall_formula             - Remove an installed formula
#   homebrew_link_formula                  - Link a formula to make binaries available

. "$MB_HELPERS_PATH/log.sh"

# Returns 0 if Homebrew is installed on the system, 1 otherwise.
# Usage:
#   homebrew_is_available
# Example:
#   if homebrew_is_available; then echo "brew found"; fi
homebrew_is_available() {
    command -v brew &> /dev/null
}

# Updates Homebrew formulae metadata to ensure latest version info is available.
# Returns 0 on success, 1 on error.
# Usage:
#   homebrew_update_metadata
# Example:
#   homebrew_update_metadata
homebrew_update_metadata() {
    if ! homebrew_is_available; then
        log debug "Homebrew não disponível, pulando atualização de metadados"
        return 0
    fi

    log debug "Atualizando formulae do Homebrew..."
    if brew update 2> /dev/null; then
        log debug "Metadados do Homebrew atualizados com sucesso"
        return 0
    else
        log debug "Falha ao atualizar metadados do Homebrew"
        return 1
    fi
}

# Returns 0 if the given Homebrew cask is installed, 1 otherwise.
# Usage:
#   homebrew_is_installed <cask_name>
# Example:
#   if homebrew_is_installed "visual-studio-code"; then echo "installed"; fi
homebrew_is_installed() {
    local cask_name="${1:-}"

    if [ -z "$cask_name" ]; then
        log error "Nome do cask é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        return 1
    fi

    brew list --cask "$cask_name" &> /dev/null
}

# Prints the installed version of a Homebrew cask, or "unknown" if not found.
# Usage:
#   homebrew_get_installed_version <cask_name>
# Example:
#   version=$(homebrew_get_installed_version "visual-studio-code")
homebrew_get_installed_version() {
    local cask_name="${1:-}"

    if [ -z "$cask_name" ]; then
        echo "unknown"
        return 0
    fi

    if ! homebrew_is_installed "$cask_name"; then
        echo "unknown"
        return 0
    fi

    local version=$(brew list --cask --versions "$cask_name" 2> /dev/null | awk '{print $2}' | head -1 || echo "unknown")
    echo "$version"
}

# Prints the latest available version of a Homebrew cask, or "unknown" if not found.
# Returns 0 if version was found, 1 on error.
# Usage:
#   homebrew_get_latest_version <cask_name>
# Example:
#   latest=$(homebrew_get_latest_version "visual-studio-code")
homebrew_get_latest_version() {
    local cask_name="${1:-}"

    if [ -z "$cask_name" ]; then
        log error "Nome do cask é obrigatório"
        echo "unknown"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado"
        echo "unknown"
        return 1
    fi

    local version=$(brew info --cask "$cask_name" 2> /dev/null | grep -m1 "^${cask_name}:" | awk '{print $2}' || echo "unknown")

    if [ "$version" = "unknown" ] || [ -z "$version" ]; then
        echo "unknown"
        return 1
    fi

    echo "$version"
    return 0
}

# Installs a cask from Homebrew. Returns 0 on success, 1 on error.
# Usage:
#   homebrew_install <cask_name> [friendly_name]
# Example:
#   homebrew_install "visual-studio-code" "VS Code"
homebrew_install() {
    local cask_name="${1:-}"
    local app_name="${2:-$cask_name}"

    if [ -z "$cask_name" ]; then
        log error "Nome do cask é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado. Por favor, instale-o primeiro:"
        log none "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        return 1
    fi

    # Check if already installed
    if homebrew_is_installed "$cask_name"; then
        local version=$(homebrew_get_installed_version "$cask_name")
        log info "$app_name $version já está instalado"
        return 0
    fi

    # Install the cask
    log info "Instalando $app_name via Homebrew..."
    log debug "Cask: $cask_name"

    if ! brew install --cask "$cask_name"; then
        log error "Falha ao instalar $app_name via Homebrew"
        return 1
    fi

    # Verify installation
    if homebrew_is_installed "$cask_name"; then
        local version=$(homebrew_get_installed_version "$cask_name")
        return 0
    else
        log error "$app_name foi instalado mas não está disponível"
        return 1
    fi
}

# Updates an installed Homebrew cask. Returns 0 if up to date or updated, 1 on error.
# Usage:
#   homebrew_update <cask_name> [friendly_name]
# Example:
#   homebrew_update "visual-studio-code" "VS Code"
homebrew_update() {
    local cask_name="${1:-}"
    local app_name="${2:-$cask_name}"

    if [ -z "$cask_name" ]; then
        log error "Nome do cask é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado"
        return 1
    fi

    # Check if installed
    if ! homebrew_is_installed "$cask_name"; then
        log error "$app_name não está instalado"
        return 1
    fi

    local current_version=$(homebrew_get_installed_version "$cask_name")
    log debug "Versão atual do $app_name: $current_version"

    # Update the cask
    if brew upgrade --cask "$cask_name" 2> /dev/null; then
        local new_version=$(homebrew_get_installed_version "$cask_name")

        if [ "$current_version" = "$new_version" ]; then
            log info "$app_name já está na versão mais recente ($new_version)"
        else
            log info "$app_name atualizado com sucesso para versão $new_version!"
        fi
        return 0
    else
        # If upgrade fails, it might already be up to date
        local new_version=$(homebrew_get_installed_version "$cask_name")
        if [ "$current_version" = "$new_version" ]; then
            log info "$app_name já está na versão mais recente ($new_version)"
            return 0
        else
            log error "Falha ao atualizar $app_name via Homebrew"
            return 1
        fi
    fi
}

# Removes an installed Homebrew cask. Returns 0 on success or if not installed, 1 on error.
# Usage:
#   homebrew_uninstall <cask_name> [friendly_name]
# Example:
#   homebrew_uninstall "visual-studio-code" "VS Code"
homebrew_uninstall() {
    local cask_name="${1:-}"
    local app_name="${2:-$cask_name}"

    if [ -z "$cask_name" ]; then
        log error "Nome do cask é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado"
        return 1
    fi

    # Check if installed
    if ! homebrew_is_installed "$cask_name"; then
        log debug "$app_name não está instalado"
        return 0
    fi

    log info "Removendo $app_name..."
    log debug "Cask: $cask_name"

    if ! brew uninstall --zap --cask "$cask_name" 2> /dev/null; then
        log error "Falha ao remover $app_name via Homebrew"
        return 1
    fi

    # Verify removal
    if ! homebrew_is_installed "$cask_name"; then
        log debug "Limpando cache de $app_name..."
        brew cleanup "$cask_name" 2> /dev/null || true
        log info "$app_name removido com sucesso!"
        return 0
    else
        log error "Falha ao remover $app_name completamente"
        return 1
    fi
}

# ============================================================================
# Formula Functions (non-cask)
# ============================================================================

# Returns 0 if the given Homebrew formula is installed, 1 otherwise.
# Usage:
#   homebrew_is_installed_formula <formula_name>
# Example:
#   if homebrew_is_installed_formula "libpq"; then echo "installed"; fi
homebrew_is_installed_formula() {
    local formula_name="${1:-}"

    if [ -z "$formula_name" ]; then
        return 1
    fi

    if ! homebrew_is_available; then
        return 1
    fi

    brew list "$formula_name" &> /dev/null
}

# Prints the installed version of a Homebrew formula, or "unknown" if not found.
# Usage:
#   homebrew_get_installed_version_formula <formula_name>
# Example:
#   version=$(homebrew_get_installed_version_formula "libpq")
homebrew_get_installed_version_formula() {
    local formula_name="${1:-}"

    if [ -z "$formula_name" ]; then
        echo "unknown"
        return 0
    fi

    if ! homebrew_is_installed_formula "$formula_name"; then
        echo "unknown"
        return 0
    fi

    local version=$(brew list --versions "$formula_name" 2> /dev/null | awk '{print $2}' | head -1 || echo "unknown")
    echo "$version"
}

# Prints the latest available version of a Homebrew formula, or "unknown" if not found.
# Returns 0 if version was found, 1 on error.
# Usage:
#   homebrew_get_latest_version_formula <formula_name>
# Example:
#   latest=$(homebrew_get_latest_version_formula "libpq")
homebrew_get_latest_version_formula() {
    local formula_name="${1:-}"

    if [ -z "$formula_name" ]; then
        echo "unknown"
        return 1
    fi

    if ! homebrew_is_available; then
        echo "unknown"
        return 1
    fi

    local version=$(brew info --json=v2 "$formula_name" 2> /dev/null | grep -oE '"stable":"[0-9]+\.[0-9]+' | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "unknown")

    if [ "$version" = "unknown" ] || [ -z "$version" ]; then
        echo "unknown"
        return 1
    fi

    echo "$version"
    return 0
}

# Installs a formula from Homebrew. Returns 0 on success, 1 on error.
# Usage:
#   homebrew_install_formula <formula_name> [friendly_name]
# Example:
#   homebrew_install_formula "libpq" "PostgreSQL client"
homebrew_install_formula() {
    local formula_name="${1:-}"
    local app_name="${2:-$formula_name}"

    if [ -z "$formula_name" ]; then
        log error "Nome da formula é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado. Por favor, instale-o primeiro:"
        log none "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        return 1
    fi

    # Check if already installed
    if homebrew_is_installed_formula "$formula_name"; then
        local version=$(homebrew_get_installed_version_formula "$formula_name")
        log info "$app_name $version já está instalado"
        return 0
    fi

    # Install the formula
    log info "Instalando $app_name via Homebrew..."
    log debug "Formula: $formula_name"

    if ! brew install "$formula_name" 2>&1 | while read -r line; do log debug "brew: $line"; done; then
        log error "Falha ao instalar $app_name via Homebrew"
        return 1
    fi

    # Verify installation
    if homebrew_is_installed_formula "$formula_name"; then
        local version=$(homebrew_get_installed_version_formula "$formula_name")
        log info "$app_name $version instalado com sucesso!"
        return 0
    else
        log error "$app_name foi instalado mas não está disponível"
        return 1
    fi
}

# Updates an installed Homebrew formula. Returns 0 if up to date or updated, 1 on error.
# Usage:
#   homebrew_update_formula <formula_name> [friendly_name]
# Example:
#   homebrew_update_formula "libpq" "PostgreSQL client"
homebrew_update_formula() {
    local formula_name="${1:-}"
    local app_name="${2:-$formula_name}"

    if [ -z "$formula_name" ]; then
        log error "Nome da formula é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado"
        return 1
    fi

    # Check if installed
    if ! homebrew_is_installed_formula "$formula_name"; then
        log error "$app_name não está instalado"
        return 1
    fi

    local current_version=$(homebrew_get_installed_version_formula "$formula_name")
    log debug "Versão atual: $current_version"

    # Update the formula
    log info "Atualizando $app_name via Homebrew..."

    if brew upgrade "$formula_name" 2>&1 | while read -r line; do log debug "brew: $line"; done; then
        local new_version=$(homebrew_get_installed_version_formula "$formula_name")

        if [ "$current_version" = "$new_version" ]; then
            log info "$app_name já estava na versão mais recente ($new_version)"
        else
            log info "$app_name atualizado com sucesso para versão $new_version!"
        fi
        return 0
    else
        # If upgrade fails, it might already be up to date
        local new_version=$(homebrew_get_installed_version_formula "$formula_name")
        if [ "$current_version" = "$new_version" ]; then
            log info "$app_name já está na versão mais recente ($new_version)"
            return 0
        else
            log error "Falha ao atualizar $app_name via Homebrew"
            return 1
        fi
    fi
}

# Removes an installed Homebrew formula. Returns 0 on success or if not installed, 1 on error.
# Usage:
#   homebrew_uninstall_formula <formula_name> [friendly_name]
# Example:
#   homebrew_uninstall_formula "libpq" "PostgreSQL client"
homebrew_uninstall_formula() {
    local formula_name="${1:-}"
    local app_name="${2:-$formula_name}"

    if [ -z "$formula_name" ]; then
        log error "Nome da formula é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado"
        return 1
    fi

    # Check if installed
    if ! homebrew_is_installed_formula "$formula_name"; then
        log debug "$app_name não está instalado"
        return 0
    fi

    log info "Removendo $app_name..."
    log debug "Formula: $formula_name"

    if ! brew uninstall "$formula_name" 2>&1 | while read -r line; do log debug "brew: $line"; done; then
        log error "Falha ao remover $app_name via Homebrew"
        return 1
    fi

    # Verify removal
    if ! homebrew_is_installed_formula "$formula_name"; then
        log debug "Limpando cache de $app_name..."
        brew cleanup "$formula_name" 2> /dev/null || true
        log info "$app_name removido com sucesso!"
        return 0
    else
        log error "Falha ao remover $app_name completamente"
        return 1
    fi
}

# Links a Homebrew formula to make its binaries available in PATH.
# Returns 0 on success, 1 on error.
# Usage:
#   homebrew_link_formula <formula_name> [force]
# Example:
#   homebrew_link_formula "libpq" "true"
homebrew_link_formula() {
    local formula_name="${1:-}"
    local force="${2:-false}"

    if [ -z "$formula_name" ]; then
        log error "Nome da formula é obrigatório"
        return 1
    fi

    if ! homebrew_is_available; then
        log error "Homebrew não está instalado"
        return 1
    fi

    if ! homebrew_is_installed_formula "$formula_name"; then
        log error "$formula_name não está instalado"
        return 1
    fi

    log debug "Criando links para $formula_name..."

    local link_args=("link" "$formula_name")
    if [ "$force" = "true" ]; then
        link_args+=("--force")
        log debug "Usando --force para sobrescrever links existentes"
    fi

    if brew "${link_args[@]}" 2>&1 | while read -r line; do log debug "brew: $line"; done; then
        log debug "Links criados com sucesso"
        return 0
    else
        log debug "Falha ao criar links"
        return 1
    fi
}
