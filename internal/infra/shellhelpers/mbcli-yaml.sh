#!/usr/bin/env bash

# Helpers para mbcli.yaml (configuração do projeto). Requer yq (Mike Farah v4).
# Carregar: . "$MB_HELPERS_PATH/mbcli-yaml.sh" ou via all.sh (após ensure.sh).
#
# Caminho: MBCLI_YAML_PATH (prioritário) ou ${MBCLI_PROJECT_ROOT:-.}/mbcli.yaml
# (raiz relativa ao PWD se MBCLI_PROJECT_ROOT não for absoluto).
# O MB CLI também lê/grava a chave `aliases` aqui com `mb alias set|unset --mbcli-yaml`.
#
# Políticas: MBCLI_YAML_ON_MISSING=error|ignore (default error)
#            MBCLI_YAML_AUTOCREATE=0|1 (default 0)

: "${MB_HELPERS_PATH:=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"

# shellcheck source=ensure.sh
. "${MB_HELPERS_PATH}/ensure.sh"

# Caminho absoluto normalizado do ficheiro alvo (stdout).
mbcli_yaml__resolved_file() {
	local f
	if [[ -n "${MBCLI_YAML_PATH:-}" ]]; then
		f="${MBCLI_YAML_PATH}"
	else
		local r="${MBCLI_PROJECT_ROOT:-.}"
		if [[ "$r" == "." ]]; then
			r="${PWD:-.}"
		elif [[ "$r" != /* ]]; then
			r="${PWD:-.}/${r}"
		fi
		f="${r%/}/mbcli.yaml"
	fi
	if [[ "$f" != /* ]]; then
		f="${PWD:-.}/${f}"
	fi
	printf '%s\n' "$f"
}

# 0 = modo ignore (ficheiro ausente não é erro nas operações acordadas).
mbcli_yaml__is_ignore_missing() {
	local m
	m="$(printf '%s' "${MBCLI_YAML_ON_MISSING:-error}" | tr '[:upper:]' '[:lower:]')"
	[[ "$m" == "ignore" ]]
}

mbcli_yaml__is_autocreate() {
	[[ "${MBCLI_YAML_AUTOCREATE:-0}" == "1" ]]
}

# Valida caminho de atribuição/remoção (deve começar por '.'; bloqueia metacaracteres óbvios).
mbcli_yaml__validate_yq_path() {
	local p="$1"
	[[ "${#p}" -ge 2 ]] || return 1
	[[ "$p" == .* ]] || return 1
	[[ "$p" != ".." ]] || return 1
	# shellcheck disable=SC2016 -- padrões literais ($(, `, etc.), não expansão do shell
	case "$p" in
	*$'\n'*|*';'*|*'&'*|*'|'*|*'>'*|*'<'*|*'$('*|*'`'*|*$'\t'*) return 1 ;;
	esac
	return 0
}

# Cria ficheiro vazio '{}' e directórios pais.
mbcli_yaml__create_empty_file() {
	local file="$1"
	local d
	d=$(dirname "$file")
	mkdir -p "$d" || return 1
	printf '%s\n' "{}" >"$file" || return 1
	return 0
}

# Aplica yq eval com argumentos restantes sobre cópia temporária e substitui o alvo.
# Uso: mbcli_yaml__atomic_apply /path/to/mbcli.yaml --arg v x '.a.b = $v'
mbcli_yaml__atomic_apply() {
	local target="$1"
	shift
	local dir tmp
	dir=$(dirname "$target")
	mkdir -p "$dir" || return 1
	tmp="${dir}/.mbcli_yaml_work.$$.$RANDOM.yml"
	if [[ -f "$target" ]]; then
		cp -p "$target" "$tmp" || return 1
	else
		printf '%s\n' "{}" >"$tmp" || return 1
	fi
	yq eval "$@" -i "$tmp" || {
		rm -f "$tmp"
		return 1
	}
	mv "$tmp" "$target" || {
		rm -f "$tmp"
		return 1
	}
	return 0
}

# Imprime o caminho resolvido de mbcli.yaml.
mbcli_yaml_path() {
	mbcli_yaml__resolved_file
}

# Garante que o ficheiro existe (cria com {} se MBCLI_YAML_AUTOCREATE=1).
mbcli_yaml_ensure() {
	local file
	file="$(mbcli_yaml__resolved_file)" || return 1
	if [[ -f "$file" ]]; then
		return 0
	fi
	if mbcli_yaml__is_autocreate; then
		mbcli_yaml__create_empty_file "$file" || return 1
		return 0
	fi
	if mbcli_yaml__is_ignore_missing; then
		return 0
	fi
	log error "mbcli_yaml_ensure: ficheiro não existe e MBCLI_YAML_AUTOCREATE não está ativo: $file"
	return 1
}

# mbcli_yaml_get <expressão_yq> — saída JSON (-o=json). Ficheiro ausente: error→1, ignore→0 e stdout vazio.
mbcli_yaml_get() {
	ensure_yq
	local file
	file="$(mbcli_yaml__resolved_file)" || return 1
	[[ -n "${1:-}" ]] || {
		log error "mbcli_yaml_get: falta a expressão yq (ex.: '.tools')"
		return 1
	}
	local filter="$1"
	if [[ ! -f "$file" ]]; then
		if mbcli_yaml__is_ignore_missing; then
			return 0
		fi
		log error "mbcli_yaml_get: ficheiro não encontrado: $file"
		return 1
	fi
	yq eval -o=json "$filter" "$file"
}

# mbcli_yaml_set <caminho_yq> <valor...> — atribuição com --arg (valor tratado como string yq).
mbcli_yaml_set() {
	ensure_yq
	local file path value
	file="$(mbcli_yaml__resolved_file)" || return 1
	path="${1:-}"
	[[ -n "$path" ]] || {
		log error "mbcli_yaml_set: falta o caminho yq (ex.: '.tools.enabled')"
		return 1
	}
	shift
	value="$*"
	mbcli_yaml__validate_yq_path "$path" || {
		log error "mbcli_yaml_set: caminho yq inválido ou inseguro: $path"
		return 1
	}

	if [[ ! -f "$file" ]]; then
		if mbcli_yaml__is_autocreate; then
			mbcli_yaml__create_empty_file "$file" || return 1
		elif mbcli_yaml__is_ignore_missing; then
			return 0
		else
			log error "mbcli_yaml_set: ficheiro não encontrado: $file"
			return 1
		fi
	fi

	# Usar strenv (sem --arg) para compatibilidade com yq mais antigos; o valor não pode conter NUL.
	export MBCLI_YAML__SETVAL="$value"
	mbcli_yaml__atomic_apply "$file" "${path} = strenv(MBCLI_YAML__SETVAL)" || {
		unset MBCLI_YAML__SETVAL
		return 1
	}
	unset MBCLI_YAML__SETVAL
	return 0
}

# mbcli_yaml_del <caminho_yq> — del(<caminho>).
mbcli_yaml_del() {
	ensure_yq
	local file path
	file="$(mbcli_yaml__resolved_file)" || return 1
	path="${1:-}"
	[[ -n "$path" ]] || {
		log error "mbcli_yaml_del: falta o caminho yq (ex.: '.tools.old')"
		return 1
	}
	mbcli_yaml__validate_yq_path "$path" || {
		log error "mbcli_yaml_del: caminho yq inválido ou inseguro: $path"
		return 1
	}

	if [[ ! -f "$file" ]]; then
		if mbcli_yaml__is_ignore_missing; then
			return 0
		fi
		log error "mbcli_yaml_del: ficheiro não encontrado: $file"
		return 1
	fi

	mbcli_yaml__atomic_apply "$file" "del(${path})" || return 1
	return 0
}
