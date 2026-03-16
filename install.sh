#!/usr/bin/env bash
set -e

# Install MB CLI, gum and glow from GitHub Releases (no sudo; installs to ~/.local/bin)

REPO="carlosdorneles-mb/mb-cli"
RELEASE_BASE="https://github.com/${REPO}/releases/download"
API_BASE="https://api.github.com/repos/${REPO}"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="mb"

GUM_REPO="charmbracelet/gum"
GUM_RELEASE_BASE="https://github.com/${GUM_REPO}/releases/download"
GUM_API_BASE="https://api.github.com/repos/${GUM_REPO}"

GLOW_REPO="charmbracelet/glow"
GLOW_RELEASE_BASE="https://github.com/${GLOW_REPO}/releases/download"
GLOW_API_BASE="https://api.github.com/repos/${GLOW_REPO}"

usage() {
  echo "Uso: $0 [--version VERSION] [--pre-release]"
  echo ""
  echo "  Baixa e instala o MB CLI, gum e glow (dependências) em ${INSTALL_DIR}."
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

# Obtém a tag da última release estável do gum (charmbracelet/gum).
get_gum_latest_tag() {
  local url="${GUM_API_BASE}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | tr -d '\n' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
  else
    echo "Requer curl para obter a versão do gum." >&2
    exit 1
  fi
}

# Obtém a tag da última release estável do glow (charmbracelet/glow).
get_glow_latest_tag() {
  local url="${GLOW_API_BASE}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | tr -d '\n' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
  else
    echo "Requer curl para obter a versão do glow." >&2
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

  # Instalar gum (dependência do MB CLI) no mesmo INSTALL_DIR
  local gum_tag gum_version GUM_OS GUM_ARCH gum_artifact gum_url_tarball gum_url_checksums gum_tmpdir gum_tarball gum_checksums_file gum_binary
  gum_tag="$(get_gum_latest_tag)"
  [ -n "$gum_tag" ] || { echo "Não foi possível obter a última versão do gum." >&2; exit 1; }
  gum_version="${gum_tag#v}"
  case "$OS" in
    linux)  GUM_OS="Linux" ;;
    darwin) GUM_OS="Darwin" ;;
    *)      echo "OS não mapeado para gum: $OS" >&2; exit 1 ;;
  esac
  case "$ARCH" in
    amd64) GUM_ARCH="x86_64" ;;
    arm64) GUM_ARCH="arm64" ;;
    *)     echo "ARCH não mapeado para gum: $ARCH" >&2; exit 1 ;;
  esac
  gum_artifact="gum_${gum_version}_${GUM_OS}_${GUM_ARCH}.tar.gz"
  gum_url_tarball="${GUM_RELEASE_BASE}/${gum_tag}/${gum_artifact}"
  gum_url_checksums="${GUM_RELEASE_BASE}/${gum_tag}/checksums.txt"
  gum_tmpdir="$(mktemp -d)"
  gum_tarball="${gum_tmpdir}/${gum_artifact}"
  gum_checksums_file="${gum_tmpdir}/checksums.txt"
  trap 'rm -rf "$tmpdir" "$gum_tmpdir"' EXIT

  echo "Baixando gum ${gum_tag} (${gum_artifact})..."
  download_file "$gum_url_tarball" "$gum_tarball" || {
    echo "Falha ao baixar gum: ${gum_url_tarball}" >&2
    exit 1
  }
  echo "Baixando checksums.txt do gum..."
  download_file "$gum_url_checksums" "$gum_checksums_file" || {
    echo "Falha ao baixar checksums.txt do gum." >&2
    exit 1
  }
  echo "Validando checksum do gum..."
  if ! verify_checksum "$gum_tarball" "$gum_checksums_file"; then
    echo "Validação do checksum do gum falhou. Instalação abortada." >&2
    exit 1
  fi
  tar -xzf "$gum_tarball" -C "$gum_tmpdir"
  gum_binary="$(find "$gum_tmpdir" -maxdepth 2 -name gum -type f | head -n1)"
  [ -n "$gum_binary" ] || { echo "Binário gum não encontrado no tarball." >&2; exit 1; }
  cp -f "$gum_binary" "${INSTALL_DIR}/gum"
  chmod +x "${INSTALL_DIR}/gum"
  echo "gum ${gum_tag} instalado em ${INSTALL_DIR}/gum"

  # Instalar glow (dependência do MB CLI) no mesmo INSTALL_DIR
  local glow_tag glow_version GLOW_OS GLOW_ARCH glow_artifact glow_url_tarball glow_url_checksums glow_tmpdir glow_tarball glow_checksums_file glow_binary
  glow_tag="$(get_glow_latest_tag)"
  [ -n "$glow_tag" ] || { echo "Não foi possível obter a última versão do glow." >&2; exit 1; }
  glow_version="${glow_tag#v}"
  case "$OS" in
    linux)  GLOW_OS="Linux" ;;
    darwin) GLOW_OS="Darwin" ;;
    *)      echo "OS não mapeado para glow: $OS" >&2; exit 1 ;;
  esac
  case "$ARCH" in
    amd64) GLOW_ARCH="x86_64" ;;
    arm64) GLOW_ARCH="arm64" ;;
    *)     echo "ARCH não mapeado para glow: $ARCH" >&2; exit 1 ;;
  esac
  glow_artifact="glow_${glow_version}_${GLOW_OS}_${GLOW_ARCH}.tar.gz"
  glow_url_tarball="${GLOW_RELEASE_BASE}/${glow_tag}/${glow_artifact}"
  glow_url_checksums="${GLOW_RELEASE_BASE}/${glow_tag}/checksums.txt"
  glow_tmpdir="$(mktemp -d)"
  glow_tarball="${glow_tmpdir}/${glow_artifact}"
  glow_checksums_file="${glow_tmpdir}/checksums.txt"
  trap 'rm -rf "$tmpdir" "$gum_tmpdir" "$glow_tmpdir"' EXIT

  echo "Baixando glow ${glow_tag} (${glow_artifact})..."
  download_file "$glow_url_tarball" "$glow_tarball" || {
    echo "Falha ao baixar glow: ${glow_url_tarball}" >&2
    exit 1
  }
  echo "Baixando checksums.txt do glow..."
  download_file "$glow_url_checksums" "$glow_checksums_file" || {
    echo "Falha ao baixar checksums.txt do glow." >&2
    exit 1
  }
  echo "Validando checksum do glow..."
  if ! verify_checksum "$glow_tarball" "$glow_checksums_file"; then
    echo "Validação do checksum do glow falhou. Instalação abortada." >&2
    exit 1
  fi
  tar -xzf "$glow_tarball" -C "$glow_tmpdir"
  glow_binary="$(find "$glow_tmpdir" -maxdepth 2 -name glow -type f | head -n1)"
  [ -n "$glow_binary" ] || { echo "Binário glow não encontrado no tarball." >&2; exit 1; }
  cp -f "$glow_binary" "${INSTALL_DIR}/glow"
  chmod +x "${INSTALL_DIR}/glow"
  echo "glow ${glow_tag} instalado em ${INSTALL_DIR}/glow"

  echo ""
  echo "MB CLI, gum e glow foram instalados em ${INSTALL_DIR}."
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
