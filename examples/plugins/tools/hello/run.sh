#!/bin/sh
# Plugin: tools/hello — exemplos de uso do gum (log, style, choose, input, confirm, spin, table, format)
. "$MB_HELPERS_PATH/all.sh"

# --- 1. Log (usa gum log via helper; respeita MB_QUIET e MB_VERBOSE) ---
log none "Olá, Mundo!"
log debug "Mensagem de debug (visível só com mb -v tools hello)"
log info "Mensagem de informação"
log warn "Aviso de exemplo"
log error "Exemplo de erro"
log fatal "Exemplo de fatal"

echo ""

# --- 2. gum style — formatação e cores (não interativo) ---
gum style --border normal --margin "0 1" --padding "0 1" --border-foreground 212 "Exemplo com gum style (borda e margem)"
gum style --foreground 86 --bold "Texto em negrito e cor"
gum style --foreground 99 --italic "Texto em itálico"
printf "%s\n" "Linha 1" "Linha 2" | gum style --border rounded --padding "0 1" --border-foreground 57

echo ""

# --- 3. gum choose — escolher uma opção ---
# Título: use --header (ou env GUM_CHOOSE_HEADER). As dicas "navigate" / "enter submit"
# vêm do binário gum e não são traduzíveis por flags; use --no-show-help para ocultá-las.
log info "Escolha uma opção (gum choose):"
OPCAO=$(gum choose --header "Escolha:" "Opção A" "Opção B" "Opção C" --no-show-help)
log info "Você escolheu: ${OPCAO:-nenhuma}"

echo ""

# --- 4. gum input — entrada de texto ---
log info "Digite um nome (gum input):"
NOME=$(gum input --placeholder "Seu nome" --no-show-help)
if [ -n "$NOME" ]; then
  gum style --foreground 212 "Olá, $NOME!" # Você poderia usar log info "Olá, $NOME!"
fi

echo ""

# --- 5. gum confirm — sim/não ---
log info "Confirmar com gum confirm:"
if gum confirm "Deseja continuar?" --default=true --affirmative "Sim" --negative "Não" --no-show-help; then
  log info "Você confirmou."
else
  log warn "Você cancelou."
fi

echo ""

# --- 6. gum spin — spinner ---
log info "Exemplo de gum spin:"
gum spin -s line --title "Processando..." -- sleep 2
log info "Spinner concluído."

echo ""

# --- 7. gum table — tabela de dados (CSV) ---
log info "Exemplo de gum table (impressão estática):"
printf "Ana,30,São Paulo\nBob,25,Rio\nCarlos,28,Belo Horizonte\n" | gum table -c Nome,Idade,Cidade -s "," -p

echo ""

# --- 8. gum format — markdown, código, etc. ---
log info "Exemplo de gum format (markdown):"
printf "%s\n" "# Exemplo Markdown" "" "**negrito** e *itálico*." "" "- item A" "- item B" | gum format
log info "Exemplo de gum format (código):"
echo 'fmt.Println("Olá, Mundo!")' | gum format -t code -l go
