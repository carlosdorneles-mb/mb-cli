#!/bin/bash

# MB CLI invocation context (see MB_CTX_* variables set when mb runs a plugin).
# Ex.: . "$MB_HELPERS_PATH/context.sh"

# mb_context_dump prints known MB_CTX_* context variables for debugging.
mb_context_dump() {
	local v
	for v in MB_CTX_INVOCATION MB_CTX_CONFIG_DIR MB_CTX_COMMAND_PATH MB_CTX_COMMAND_NAME MB_CTX_PARENT_COMMAND_PATH MB_CTX_COBR_COMMAND_PATH MB_CTX_PLUGIN_FLAGS MB_CTX_PEER_COMMANDS MB_CTX_CHILD_COMMANDS MB_CTX_CHILD_COMMAND_INFO MB_CTX_HIDDEN_CHILD_COMMANDS MB_CTX_CHILD_COMMAND_ALIASES; do
		printf '%s=%s\n' "$v" "${!v-}"
	done
}

# mb_context_dump_json prints one JSON object with the same MB_CTX_* fields (snake_case keys).
# Prefers jq; falls back to python3 -c json. Writes to stdout; returns 1 if neither is available.
mb_context_dump_json() {
	local peers="${MB_CTX_PEER_COMMANDS:-[]}"
	local child="${MB_CTX_CHILD_COMMANDS:-[]}"
	local child_info="${MB_CTX_CHILD_COMMAND_INFO:-[]}"
	local hidden="${MB_CTX_HIDDEN_CHILD_COMMANDS:-[]}"
	local aliases="${MB_CTX_CHILD_COMMAND_ALIASES:-[]}"
	if command -v jq >/dev/null 2>&1; then
		jq -n \
			--arg invocation "${MB_CTX_INVOCATION-}" \
			--arg config_dir "${MB_CTX_CONFIG_DIR-}" \
			--arg command_path "${MB_CTX_COMMAND_PATH-}" \
			--arg command_name "${MB_CTX_COMMAND_NAME-}" \
			--arg parent_command_path "${MB_CTX_PARENT_COMMAND_PATH-}" \
			--arg cobr_command_path "${MB_CTX_COBR_COMMAND_PATH-}" \
			--arg plugin_flags_str "${MB_CTX_PLUGIN_FLAGS-}" \
			--arg peer_json "$peers" \
			--arg child_json "$child" \
			--arg child_info_json "$child_info" \
			--arg hidden_json "$hidden" \
			--arg aliases_json "$aliases" \
			'{
				invocation: $invocation,
				config_dir: $config_dir,
				command_path: $command_path,
				command_name: $command_name,
				parent_command_path: $parent_command_path,
				cobr_command_path: $cobr_command_path,
				plugin_flags: (if $plugin_flags_str == "" then [] else ($plugin_flags_str | split(" ") | map(select(length > 0))) end),
				peer_commands: (try ($peer_json | fromjson) catch []),
				child_commands: (try ($child_json | fromjson) catch []),
				child_command_info: (try ($child_info_json | fromjson) catch []),
				hidden_child_commands: (try ($hidden_json | fromjson) catch []),
				child_command_aliases: (try ($aliases_json | fromjson) catch [])
			}'
		return 0
	fi
	if command -v python3 >/dev/null 2>&1; then
		python3 - <<'PY'
import json, os
flags = os.environ.get("MB_CTX_PLUGIN_FLAGS", "")
parts = [x for x in flags.split() if x]

def loads_arr(key):
    raw = os.environ.get(key, "[]")
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return []

peers = loads_arr("MB_CTX_PEER_COMMANDS")
child = loads_arr("MB_CTX_CHILD_COMMANDS")
child_info = loads_arr("MB_CTX_CHILD_COMMAND_INFO")
hidden = loads_arr("MB_CTX_HIDDEN_CHILD_COMMANDS")
aliases = loads_arr("MB_CTX_CHILD_COMMAND_ALIASES")
obj = {
    "invocation": os.environ.get("MB_CTX_INVOCATION", ""),
    "config_dir": os.environ.get("MB_CTX_CONFIG_DIR", ""),
    "command_path": os.environ.get("MB_CTX_COMMAND_PATH", ""),
    "command_name": os.environ.get("MB_CTX_COMMAND_NAME", ""),
    "parent_command_path": os.environ.get("MB_CTX_PARENT_COMMAND_PATH", ""),
    "cobr_command_path": os.environ.get("MB_CTX_COBR_COMMAND_PATH", ""),
    "plugin_flags": parts,
    "peer_commands": peers,
    "child_commands": child,
    "child_command_info": child_info,
    "hidden_child_commands": hidden,
    "child_command_aliases": aliases,
}
print(json.dumps(obj, ensure_ascii=False))
PY
		return 0
	fi
	echo "mb_context_dump_json: precisa de jq ou python3 no PATH" >&2
	return 1
}

# mb_peer_commands_lines prints one sibling command name per line from MB_CTX_PEER_COMMANDS (JSON array).
# Uses jq when available; otherwise a best-effort parse for simple names (letters, digits, _, -).
mb_peer_commands_lines() {
	local j="${MB_CTX_PEER_COMMANDS:-[]}"
	if command -v jq >/dev/null 2>&1; then
		jq -r '.[]' <<<"$j" 2>/dev/null
		return
	fi
	j="${j#\[}"
	j="${j%\]}"
	j="${j//\"/}"
	local IFS=,
	local part
	for part in $j; do
		part="${part// /}"
		[[ -n "$part" ]] && printf '%s\n' "$part"
	done
}

# mb_child_commands_lines — um nome por linha a partir de MB_CTX_CHILD_COMMANDS (filhos Cobra não ocultos).
mb_child_commands_lines() {
	local j="${MB_CTX_CHILD_COMMANDS:-[]}"
	if command -v jq >/dev/null 2>&1; then
		jq -r '.[]' <<<"$j" 2>/dev/null
		return
	fi
	j="${j#\[}"
	j="${j%\]}"
	j="${j//\"/}"
	local IFS=,
	local part
	for part in $j; do
		part="${part// /}"
		[[ -n "$part" ]] && printf '%s\n' "$part"
	done
}

# mb_child_command_info_json — imprime MB_CTX_CHILD_COMMAND_INFO (JSON `[{name,description},…]`); vazio → `[]`.
mb_child_command_info_json() {
	printf '%s\n' "${MB_CTX_CHILD_COMMAND_INFO:-[]}"
}

# mb_hidden_child_commands_lines — um nome por linha a partir de MB_CTX_HIDDEN_CHILD_COMMANDS.
mb_hidden_child_commands_lines() {
	local j="${MB_CTX_HIDDEN_CHILD_COMMANDS:-[]}"
	if command -v jq >/dev/null 2>&1; then
		jq -r '.[]' <<<"$j" 2>/dev/null
		return
	fi
	j="${j#\[}"
	j="${j%\]}"
	j="${j//\"/}"
	local IFS=,
	local part
	for part in $j; do
		part="${part// /}"
		[[ -n "$part" ]] && printf '%s\n' "$part"
	done
}

# mb_ctx_has_plugin_flag NAME — exit 0 if NAME is among MB_CTX_PLUGIN_FLAGS (nome longo).
mb_ctx_has_plugin_flag() {
	local want="$1"
	[[ -n "$want" ]] || return 1
	local f
	for f in ${MB_CTX_PLUGIN_FLAGS-}; do
		[[ "$f" == "$want" ]] && return 0
	done
	return 1
}

# mb_ctx_peer_contains NAME — exit 0 if NAME appears in MB_CTX_PEER_COMMANDS (como irmão).
mb_ctx_peer_contains() {
	local want="$1"
	[[ -n "$want" ]] || return 1
	local peer
	while IFS= read -r peer; do
		[[ "$peer" == "$want" ]] && return 0
	done < <(mb_peer_commands_lines)
	return 1
}

# mb_ctx_child_contains NAME — exit 0 if NAME appears in MB_CTX_CHILD_COMMANDS (filho directo visível).
mb_ctx_child_contains() {
	local want="$1"
	[[ -n "$want" ]] || return 1
	local name
	while IFS= read -r name; do
		[[ "$name" == "$want" ]] && return 0
	done < <(mb_child_commands_lines)
	return 1
}

# mb_ctx_hidden_child_contains NAME — exit 0 if NAME appears em MB_CTX_HIDDEN_CHILD_COMMANDS.
mb_ctx_hidden_child_contains() {
	local want="$1"
	[[ -n "$want" ]] || return 1
	local name
	while IFS= read -r name; do
		[[ "$name" == "$want" ]] && return 0
	done < <(mb_hidden_child_commands_lines)
	return 1
}

# mb_ctx_peer_count — imprime o número de comandos irmãos (inteiro, uma linha).
mb_ctx_peer_count() {
	local n=0
	while IFS= read -r _; do
		((++n)) || true
	done < <(mb_peer_commands_lines)
	printf '%s\n' "$n"
}

# mb_ctx_child_count — número de filhos Cobra visíveis (não ocultos).
mb_ctx_child_count() {
	local n=0
	while IFS= read -r _; do
		((++n)) || true
	done < <(mb_child_commands_lines)
	printf '%s\n' "$n"
}

# mb_ctx_hidden_child_count — número de filhos Cobra ocultos.
mb_ctx_hidden_child_count() {
	local n=0
	while IFS= read -r _; do
		((++n)) || true
	done < <(mb_hidden_child_commands_lines)
	printf '%s\n' "$n"
}

# mb_ctx_cache_db — imprime o caminho de cache.db sob MB_CTX_CONFIG_DIR; exit 1 se o dir estiver vazio.
mb_ctx_cache_db() {
	[[ -n "${MB_CTX_CONFIG_DIR:-}" ]] || return 1
	printf '%s\n' "${MB_CTX_CONFIG_DIR}/cache.db"
}

# mb_ctx_parent_is PATH — exit 0 se MB_CTX_PARENT_COMMAND_PATH for exatamente PATH.
mb_ctx_parent_is() {
	[[ "${MB_CTX_PARENT_COMMAND_PATH-}" == "$1" ]]
}

# mb_ctx_command_path_is PATH — exit 0 se MB_CTX_COMMAND_PATH for exatamente PATH (manifest).
mb_ctx_command_path_is() {
	[[ "${MB_CTX_COMMAND_PATH-}" == "$1" ]]
}

# mb_ctx_path_depth — imprime o número de segmentos de MB_CTX_COMMAND_PATH (ex.: tools/vscode → 2; hello → 1; vazio → 0).
mb_ctx_path_depth() {
	local p="${MB_CTX_COMMAND_PATH-}"
	if [[ -z "$p" ]]; then
		printf '0\n'
		return 0
	fi
	local IFS=/
	# shellcheck disable=SC2086
	set -- $p
	printf '%s\n' "$#"
}
