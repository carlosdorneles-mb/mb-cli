#!/bin/bash
# Managed blocks in ~/.bashrc and ~/.zshrc using mb-cli-plugins markers.
# Only modifies files that already exist. Idempotent: does not append again if the begin marker line is already present.

# shell_rc_ensure_block MARKER_BEGIN MARKER_END BODY
# For each existing rc file, appends: blank line + MARKER_BEGIN + BODY + MARKER_END (if MARKER_BEGIN is not already found).
shell_rc_ensure_block() {
    local marker_begin="$1"
    local marker_end="$2"
    local body="$3"
    local f

    if [ -z "$marker_begin" ] || [ -z "$marker_end" ]; then
        return 1
    fi

    for f in "${HOME}/.bashrc" "${HOME}/.zshrc"; do
        [ -f "$f" ] || continue
        if grep -Fq "$marker_begin" "$f" 2>/dev/null; then
            continue
        fi
        {
            printf '\n%s\n' "$marker_begin"
            printf '%s\n' "$body"
            printf '%s\n' "$marker_end"
        } >> "$f" || return 1
    done
}

# shell_rc_remove_block MARKER_BEGIN MARKER_END
# Removes lines from MARKER_BEGIN through MARKER_END (inclusive) in each existing rc that contains the block.
shell_rc_remove_block() {
    local marker_begin="$1"
    local marker_end="$2"
    local f tmp

    if [ -z "$marker_begin" ] || [ -z "$marker_end" ]; then
        return 1
    fi

    for f in "${HOME}/.bashrc" "${HOME}/.zshrc"; do
        [ -f "$f" ] || continue
        grep -Fq "$marker_begin" "$f" 2>/dev/null || continue
        tmp=$(mktemp) || return 1
        if awk -v b="$marker_begin" -v e="$marker_end" '
            $0 == b { skip = 1; next }
            $0 == e { skip = 0; next }
            !skip { print }
        ' "$f" > "$tmp"; then
            mv "$tmp" "$f" || {
                rm -f "$tmp"
                return 1
            }
        else
            rm -f "$tmp"
            return 1
        fi
    done
}
