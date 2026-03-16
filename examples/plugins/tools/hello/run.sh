#!/bin/sh
# Plugin: tools/hello
. "$MB_HELPERS_PATH"

log info "Olá, Mundo!"
log debug "Mensagem de debug (visível só com mb -v tools hello)"
log info "Oi essa é uma mensagem de informação indo para o gum log"
log warn "Aviso de exemplo"
log error "Exemplo de erro"
log fatal "Exemplo de fatal"
