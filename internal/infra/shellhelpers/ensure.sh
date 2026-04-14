#!/bin/bash

# log.sh (via all.sh normalmente já carregado)
: "${MB_HELPERS_PATH:=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"

# shellcheck source=log.sh
. "${MB_HELPERS_PATH}/log.sh"

ensure_npx() {
    if ! command -v npx >/dev/null 2>&1; then
        log error "Comando 'npx' não encontrado. Instale Node.js (inclui npm/npx): https://nodejs.org/"
        exit 1
    fi
}

ensure_jq() {
    if ! command -v jq >/dev/null 2>&1; then
        log error "Comando 'jq' não encontrado. Instale: https://stedolan.github.io/jq/"
        exit 1
    fi
}

ensure_yq() {
    if ! command -v yq >/dev/null 2>&1; then
        log error "Comando 'yq' não encontrado. Instale: https://github.com/mikefarah/yq"
        exit 1
    fi
}

ensure_gum() {
    if ! command -v gum >/dev/null 2>&1; then
        log error "Comando 'gum' não encontrado. Instale: https://github.com/charmbracelet/gum"
        exit 1
    fi
}

ensure_fzf() {
    if ! command -v fzf >/dev/null 2>&1; then
        log error "Comando 'fzf' não encontrado. Instale: https://github.com/junegunn/fzf"
        exit 1
    fi
}

ensure_kubectl() {
    if ! command -v kubectl >/dev/null 2>&1; then
        log error "Comando 'kubectl' não encontrado. Instale: https://kubernetes.io/docs/tasks/tools/install-kubectl/"
        exit 1
    fi
}