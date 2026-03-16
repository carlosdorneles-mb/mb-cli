#!/usr/bin/env bash
set -e

# Remove MB CLI binary (installed in ~/.local/bin by install.sh)

INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="mb"

bin="${INSTALL_DIR}/${BINARY_NAME}"
if [ -f "$bin" ]; then
  rm -f "$bin"
  echo "MB CLI removido (${bin})"
else
  echo "Binário não encontrado: ${bin}" >&2
  exit 1
fi

echo "Os dados do CLI (plugins, config) permanecem em ~/.config/mb (Linux) ou ~/Library/Application Support/mb (macOS)."
echo "gum, glow, jq e fzf (se instalados pelo install.sh) permanecem em ${INSTALL_DIR}."
