#!/usr/bin/env bash
set -e

# Install MB CLI from GitHub Releases (no sudo; installs to ~/.local/bin)

REPO="carlosdorneles-mb/mb-cli"
RELEASE_BASE="https://github.com/${REPO}/releases/download"
API_BASE="https://api.github.com/repos/${REPO}"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="mb"

usage() {
  echo "Uso: $0 [--version VERSION] [--pre-release]"
  echo ""
  echo "  Baixa e instala o MB CLI em ${INSTALL_DIR}."
  echo "  Sem opções     Usa a última versão estável (consulta a API do GitHub)."
  echo "  --version N    Usa a versão N (ex.: 0.0.5 ou v0.0.5)."
  echo "  --pre-release  Usa a última versão pre-release (requer jq)."
  echo ""
  echo "Para remover: execute uninstall.sh"
  echo "Requer: curl ou wget. Instalação é validada com checksums.txt do release."
  exit 1
}

# Obtém a tag da última versão estável (API: /releases/latest).
get_latest_tag() {
  local url="${API_BASE}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | tr -d '\n' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
  else
    echo "Requer curl para obter a última versão." >&2
    exit 1
  fi
}

# Obtém a tag da última versão pre-release (requer jq).
get_latest_prerelease_tag() {
  local url="${API_BASE}/releases"
  if ! command -v jq >/dev/null 2>&1; then
    echo "Requer jq para --pre-release. Instale jq ou use --version X.Y.Z" >&2
    exit 1
  fi
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | jq -r '[.[] | select(.prerelease==true)][0].tag_name // empty'
  else
    echo "Requer curl para obter pre-release." >&2
    exit 1
  fi
}

detect_os_arch() {
  local uname_s uname_m
  uname_s="$(uname -s)"
  uname_m="$(uname -m)"

  case "$uname_s" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)
      echo "Sistema operacional não suportado: $uname_s" >&2
      exit 1
      ;;
  esac

  case "$uname_m" in
    x86_64|amd64)   ARCH="amd64" ;;
    aarch64|arm64)  ARCH="arm64" ;;
    *)
      echo "Arquitetura não suportada: $uname_m" >&2
      exit 1
      ;;
  esac
}

download_file() {
  local url="$1" dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -sSLf -o "$dest" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$dest" "$url"
  else
    echo "Requer curl ou wget para download." >&2
    exit 1
  fi
}

verify_checksum() {
  local tarball="$1" checksums_file="$2" artifact_name
  artifact_name="$(basename "$tarball")"

  local onefile
  onefile="$(dirname "$tarball")/checksum.one"
  grep -F "$artifact_name" "$checksums_file" > "$onefile" || true
  if [ ! -s "$onefile" ]; then
    echo "Checksum não encontrado para $artifact_name em checksums.txt" >&2
    return 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$(dirname "$tarball")" && sha256sum -c "$(basename "$onefile")")
  elif command -v shasum >/dev/null 2>&1; then
    (cd "$(dirname "$tarball")" && shasum -a 256 -c "$(basename "$onefile")")
  else
    echo "Requer sha256sum (Linux) ou shasum (macOS) para validar o download." >&2
    return 1
  fi
}

do_install() {
  local version="$1" tag url_tarball url_checksums tmpdir tarball checksums_file
  version="${version#v}"
  tag="v${version}"
  detect_os_arch

  local artifact="${BINARY_NAME}_${version}_${OS}_${ARCH}.tar.gz"
  url_tarball="${RELEASE_BASE}/${tag}/${artifact}"
  url_checksums="${RELEASE_BASE}/${tag}/checksums.txt"

  tmpdir="$(mktemp -d)"
  tarball="${tmpdir}/${artifact}"
  checksums_file="${tmpdir}/checksums.txt"

  trap 'rm -rf "$tmpdir"' EXIT

  echo "Baixando ${artifact}..."
  download_file "$url_tarball" "$tarball" || {
    echo "Falha ao baixar ${url_tarball}" >&2
    exit 1
  }
  echo "Baixando checksums.txt..."
  download_file "$url_checksums" "$checksums_file" || {
    echo "Falha ao baixar checksums.txt" >&2
    exit 1
  }

  echo "Validando checksum..."
  if ! verify_checksum "$tarball" "$checksums_file"; then
    echo "Validação do checksum falhou. Instalação abortada." >&2
    exit 1
  fi

  mkdir -p "$INSTALL_DIR"
  tar -xzf "$tarball" -C "$tmpdir"
  cp -f "${tmpdir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  echo "MB CLI ${tag} instalado em ${INSTALL_DIR}/${BINARY_NAME}"
  if ! echo "$PATH" | grep -qF "${INSTALL_DIR}"; then
    echo "Adicione ${INSTALL_DIR} ao seu PATH, por exemplo em ~/.profile ou ~/.bashrc:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
  echo "Depois rode: mb self sync"
}

# Parse arguments
VERSION=""
USE_PRERELEASE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --version)
      [ $# -gt 1 ] || usage
      VERSION="$2"
      shift 2
      ;;
    --pre-release)
      USE_PRERELEASE=1
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage
      ;;
  esac
done

if [ -n "$USE_PRERELEASE" ]; then
  VERSION="$(get_latest_prerelease_tag)"
  [ -n "$VERSION" ] || { echo "Nenhum pre-release encontrado." >&2; exit 1; }
elif [ -z "$VERSION" ]; then
  VERSION="$(get_latest_tag)"
  [ -n "$VERSION" ] || { echo "Não foi possível obter a última versão." >&2; exit 1; }
fi

do_install "$VERSION"
