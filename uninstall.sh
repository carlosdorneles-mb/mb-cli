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

gum_bin="${INSTALL_DIR}/gum"
if [ -f "$gum_bin" ]; then
  rm -f "$gum_bin"
  echo "gum removido (${gum_bin}) — havia sido instalado pelo install.sh."
fi

glow_bin="${INSTALL_DIR}/glow"
if [ -f "$glow_bin" ]; then
  rm -f "$glow_bin"
  echo "glow removido (${glow_bin}) — havia sido instalado pelo install.sh."
fi

echo "Os dados do CLI (plugins, config) permanecem em ~/.config/mb (Linux) ou ~/Library/Application Support/mb (macOS)."
