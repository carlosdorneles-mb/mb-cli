#!/bin/sh
# Exemplo: executado quando o usuário roda mb tools do --rollback
. "${MB_HELPERS_PATH:-.}/all.sh" 2>/dev/null || true
log info "Rollback (simulado)."
echo "Você rodou: mb tools do --rollback"
