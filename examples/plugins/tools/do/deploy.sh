#!/bin/sh
# Exemplo: executado quando o usuário roda mb tools do --deploy
. "${MB_HELPERS_PATH:-.}/all.sh" 2>/dev/null || true
log info "Deploy (simulado)."
echo "Você rodou: mb tools do --deploy"
