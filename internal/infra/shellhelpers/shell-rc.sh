#!/bin/bash
# Blocos geridos por ferramentas MB (mb tools …): vivem em ~/.config/mb/shell-extras.sh
# entre marcadores # mb-cli-plugins:<nome>:begin / :end.
#
# ~/.bashrc e ~/.zshrc (se existirem) recebem no máximo um stub que faz source desse ficheiro.
# Migração: ao garantir um bloco, se o marcador ainda estiver só nos RCs legados, o bloco é
# removido dali e passa a existir apenas em shell-extras.sh (evita duplicar init).

# Caminho absoluto de shell-extras.sh (não cria directórios).
shell_rc_extras_path() {
	printf '%s\n' "${XDG_CONFIG_HOME:-$HOME/.config}/mb/shell-extras.sh"
}

# Garante o directório ~/.config/mb e um shell-extras.sh inicial se faltar.
shell_rc_touch_extras_file() {
	local p dir
	p="$(shell_rc_extras_path)"
	dir="$(dirname "$p")"
	mkdir -p "$dir" || return 1
	if [ ! -f "$p" ]; then
		{
			printf '%s\n' '# shell-extras.sh — carregado pelos stubs em ~/.bashrc e ~/.zshrc (MB CLI).'
			printf '%s\n' '# Blocos entre # mb-cli-plugins:…:begin/end são geridos por mb tools install/uninstall.'
			printf '\n'
		} >"$p" || return 1
	fi
}

# Acrescenta em ~/.bashrc e ~/.zshrc existentes uma linha que faz source de shell-extras.sh (idempotente).
shell_rc_ensure_stub() {
	local stub_line='[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/mb/shell-extras.sh" ] && . "${XDG_CONFIG_HOME:-$HOME/.config}/mb/shell-extras.sh"'
	local f
	for f in "${HOME}/.bashrc" "${HOME}/.zshrc"; do
		[ -f "$f" ] || continue
		if grep -Fq "mb/shell-extras.sh" "$f" 2>/dev/null; then
			continue
		fi
		{
			printf '\n%s\n' '# MB CLI — carrega ~/.config/mb/shell-extras.sh (ferramentas mb tools)'
			printf '%s\n' "$stub_line"
		} >>"$f" || return 1
	done
}

# shell_rc__remove_block_from_file FILE MARKER_BEGIN MARKER_END — remove bloco inclusivo; ficheiro inalterado se marcador ausente.
shell_rc__remove_block_from_file() {
	local f="$1" marker_begin="$2" marker_end="$3" tmp

	[ -f "$f" ] || return 0
	grep -Fq "$marker_begin" "$f" 2>/dev/null || return 0

	tmp=$(mktemp) || return 1
	if awk -v b="$marker_begin" -v e="$marker_end" '
		$0 == b { skip = 1; next }
		$0 == e { skip = 0; next }
		!skip { print }
	' "$f" >"$tmp"; then
		mv "$tmp" "$f" || {
			rm -f "$tmp"
			return 1
		}
	else
		rm -f "$tmp"
		return 1
	fi
}

# shell_rc_ensure_block MARKER_BEGIN MARKER_END BODY
# Garante stub + shell-extras.sh; acrescenta bloco só em shell-extras.sh se o begin não existir lá.
# Se o bloco existir só em ~/.bashrc ou ~/.zshrc (legado), remove-o dali antes de acrescentar em extras.
shell_rc_ensure_block() {
	local marker_begin="$1"
	local marker_end="$2"
	local body="$3"
	local extras f

	if [ -z "$marker_begin" ] || [ -z "$marker_end" ]; then
		return 1
	fi

	shell_rc_touch_extras_file || return 1
	shell_rc_ensure_stub || return 1

	extras="$(shell_rc_extras_path)"
	if grep -Fq "$marker_begin" "$extras" 2>/dev/null; then
		return 0
	fi

	for f in "${HOME}/.bashrc" "${HOME}/.zshrc"; do
		if [ -f "$f" ] && grep -Fq "$marker_begin" "$f" 2>/dev/null; then
			shell_rc__remove_block_from_file "$f" "$marker_begin" "$marker_end" || return 1
		fi
	done

	{
		printf '\n%s\n' "$marker_begin"
		printf '%s\n' "$body"
		printf '%s\n' "$marker_end"
	} >>"$extras" || return 1
}

# shell_rc_remove_block MARKER_BEGIN MARKER_END
# Remove o bloco de shell-extras.sh (se existir) e dos RCs legados.
shell_rc_remove_block() {
	local marker_begin="$1"
	local marker_end="$2"
	local extras

	if [ -z "$marker_begin" ] || [ -z "$marker_end" ]; then
		return 1
	fi

	extras="$(shell_rc_extras_path)"
	for f in "$extras" "${HOME}/.bashrc" "${HOME}/.zshrc"; do
		[ -f "$f" ] || continue
		shell_rc__remove_block_from_file "$f" "$marker_begin" "$marker_end" || return 1
	done
}
