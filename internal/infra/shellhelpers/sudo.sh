#!/bin/bash

# Check if the script is running with superuser privileges
# Returns:
#   0 - If running as root
#   1 - If not running with root privileges
check_sudo() {
    if [ "$EUID" -ne 0 ]; then
        log warn "Este comando requer privilégios de superusuário (sudo)." >&2
        return 1
    fi
    return 0
}

# Ensure the script has superuser privileges
# If not running as root, requests sudo authentication
# Exits with error if authentication fails
# Args:
#   $@ - Command arguments (checked for bypass)
required_sudo() {
    if ! check_sudo; then
        sudo -v || {
            log error "Falha ao obter privilégios de sudo"
            exit 1
        }
    fi
}
