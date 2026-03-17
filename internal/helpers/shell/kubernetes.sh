#!/bin/sh

. "$MB_HELPERS_PATH/log.sh"

# Checks if kubectl is installed. Prints an error and exits if not found and "exit_on_error" is passed as an argument.
# Usage:
#   kb_check_installed [exit_on_error]
# Example:
#   kb_check_installed "exit_on_error"
kb_check_installed() {
    if ! command -v kubectl &> /dev/null; then
        log error "kubectl is not installed. See: https://kubernetes.io/docs/tasks/tools/install-kubectl/"
        if [ "$1" == "exit_on_error" ]; then
            exit 1
        fi
        return 1
    fi
    return 0
}

# Checks if a given Kubernetes namespace exists. Exits with error if not found and "exit_on_error" is passed as the second argument.
# Usage:
#   kb_check_namespace_exists <namespace> [exit_on_error]
# Example:
#   kb_check_namespace_exists "production" "exit_on_error"
kb_check_namespace_exists() {
    local namespace=$1
    kb_check_installed "exit_on_error"

    if ! kubectl get namespace "$namespace" &> /dev/null; then
        if [ "$2" == "exit_on_error" ]; then
            log error "Namespace '$namespace' not found."
            exit 1
        fi
        return 1
    fi
    return 0
}

# Returns the current kubectl context name.
# Usage:
#   kb_get_current_context
# Example:
#   context=$(kb_get_current_context)
#   echo "$context"
kb_get_current_context() {
    kb_check_installed "exit_on_error"
    kubectl config current-context
}

# Prints the current kubectl context to the console.
# Usage:
#   kb_print_current_context
# Example:
#   kb_print_current_context
kb_print_current_context() {
    local context
    context=$(kb_get_current_context)
    if [ $? -eq 0 ]; then
        echo "Current kubectl context: $context"
    else
        log error "Could not retrieve the current kubectl context."
    fi
}
