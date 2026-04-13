#!/usr/bin/env bash
set -e

# Install MB CLI, gum, glow, jq and fzf from GitHub Releases (no sudo; installs to ~/.local/bin).
# No macOS, se o Homebrew estiver disponível, instala também o mas (CLI da Mac App Store).

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

JQ_REPO="jqlang/jq"
JQ_RELEASE_BASE="https://github.com/${JQ_REPO}/releases/download"
JQ_API_BASE="https://api.github.com/repos/${JQ_REPO}"

FZF_REPO="junegunn/fzf"
FZF_RELEASE_BASE="https://github.com/${FZF_REPO}/releases/download"
FZF_API_BASE="https://api.github.com/repos/${FZF_REPO}"

YQ_REPO="mikefarah/yq"
YQ_RELEASE_BASE="https://github.com/${YQ_REPO}/releases/download"
YQ_API_BASE="https://api.github.com/repos/${YQ_REPO}"

usage() {
  echo "Uso: $0 [--version VERSION] [--pre-release]"
  echo ""
  echo "  Baixa e instala o MB CLI, gum, glow, jq, fzf e yq (dependências) em ${INSTALL_DIR}."
  echo "  No macOS com Homebrew: dependências instaladas via brew (gum, glow, jq, fzf, yq, mas)."
  echo "  Sem opções     Usa a última versão estável (consulta a API do GitHub)."
  echo "  --version N    Usa a versão N (ex.: 0.0.5 ou v0.0.5)."
  echo "  --pre-release  Usa a última versão pre-release (requer jq)."
  echo ""
  echo "Para remover: execute uninstall.sh"
  echo "Requer: curl ou wget. Apenas a instalação do MB CLI é validada com checksums.txt do release."
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

# Obtém a tag da última release estável do jq (jqlang/jq). Tag vem como jq-1.8.1.
get_jq_latest_tag() {
  local url="${JQ_API_BASE}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | tr -d '\n' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
  else
    echo "Requer curl para obter a versão do jq." >&2
    exit 1
  fi
}

# Obtém a tag da última release estável do fzf (junegunn/fzf).
get_fzf_latest_tag() {
  local url="${FZF_API_BASE}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | tr -d '\n' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
  else
    echo "Requer curl para obter a versão do fzf." >&2
    exit 1
  fi
}

# Obtém a tag da última release estável do yq (mikefarah/yq).
get_yq_latest_tag() {
  local url="${YQ_API_BASE}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" | tr -d '\n' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
  else
    echo "Requer curl para obter a versão do yq." >&2
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

# --- Funções de instalação manual de dependências (Linux / sem Homebrew) ---

install_gum_manual() {
  local gum_tag gum_version GUM_OS GUM_ARCH gum_artifact gum_url_tarball gum_tmpdir gum_tarball gum_binary
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
  gum_tmpdir="$(mktemp -d)"
  gum_tarball="${gum_tmpdir}/${gum_artifact}"
  trap 'rm -rf "$tmpdir" "$gum_tmpdir" "${glow_tmpdir:-}" "${jq_tmpdir:-}" "${fzf_tmpdir:-}"' EXIT

  echo "Baixando gum ${gum_tag} (${gum_artifact})..."
  download_file "$gum_url_tarball" "$gum_tarball" || {
    echo "Falha ao baixar gum: ${gum_url_tarball}" >&2
    exit 1
  }
  tar -xzf "$gum_tarball" -C "$gum_tmpdir"
  gum_binary="$(find "$gum_tmpdir" -maxdepth 2 -name gum -type f | head -n1)"
  [ -n "$gum_binary" ] || { echo "Binário gum não encontrado no tarball." >&2; exit 1; }
  cp -f "$gum_binary" "${INSTALL_DIR}/gum"
  chmod +x "${INSTALL_DIR}/gum"
  echo "gum ${gum_tag} instalado em ${INSTALL_DIR}/gum"
}

install_glow_manual() {
  local glow_tag glow_version GLOW_OS GLOW_ARCH glow_artifact glow_url_tarball glow_tmpdir glow_tarball glow_binary
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
  glow_tmpdir="$(mktemp -d)"
  glow_tarball="${glow_tmpdir}/${glow_artifact}"
  trap 'rm -rf "$tmpdir" "$gum_tmpdir" "$glow_tmpdir" "${jq_tmpdir:-}" "${fzf_tmpdir:-}"' EXIT

  echo "Baixando glow ${glow_tag} (${glow_artifact})..."
  download_file "$glow_url_tarball" "$glow_tarball" || {
    echo "Falha ao baixar glow: ${glow_url_tarball}" >&2
    exit 1
  }
  tar -xzf "$glow_tarball" -C "$glow_tmpdir"
  glow_binary="$(find "$glow_tmpdir" -maxdepth 2 -name glow -type f | head -n1)"
  [ -n "$glow_binary" ] || { echo "Binário glow não encontrado no tarball." >&2; exit 1; }
  cp -f "$glow_binary" "${INSTALL_DIR}/glow"
  chmod +x "${INSTALL_DIR}/glow"
  echo "glow ${glow_tag} instalado em ${INSTALL_DIR}/glow"
}

install_jq_manual() {
  local jq_tag jq_os jq_arch jq_artifact jq_url jq_tmpdir
  jq_tag="$(get_jq_latest_tag)"
  [ -n "$jq_tag" ] || { echo "Não foi possível obter a última versão do jq." >&2; exit 1; }
  case "$OS" in
    linux)  jq_os="linux" ;;
    darwin) jq_os="macos" ;;
    *)      echo "OS não mapeado para jq: $OS" >&2; exit 1 ;;
  esac
  jq_arch="$ARCH"
  jq_artifact="jq-${jq_os}-${jq_arch}"
  jq_url="${JQ_RELEASE_BASE}/${jq_tag}/${jq_artifact}"
  jq_tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir" "$gum_tmpdir" "$glow_tmpdir" "$jq_tmpdir" "${fzf_tmpdir:-}"' EXIT
  echo "Baixando jq ${jq_tag} (${jq_artifact})..."
  download_file "$jq_url" "${jq_tmpdir}/jq" || {
    echo "Falha ao baixar jq: ${jq_url}" >&2
    exit 1
  }
  chmod +x "${jq_tmpdir}/jq"
  cp -f "${jq_tmpdir}/jq" "${INSTALL_DIR}/jq"
  echo "jq ${jq_tag} instalado em ${INSTALL_DIR}/jq"
}

install_fzf_manual() {
  local fzf_tag fzf_version fzf_os_arch fzf_artifact fzf_url_tarball fzf_tmpdir fzf_tarball fzf_binary
  fzf_tag="$(get_fzf_latest_tag)"
  [ -n "$fzf_tag" ] || { echo "Não foi possível obter a última versão do fzf." >&2; exit 1; }
  fzf_version="${fzf_tag#v}"
  fzf_os_arch="${OS}_${ARCH}"
  fzf_artifact="fzf-${fzf_version}-${fzf_os_arch}.tar.gz"
  fzf_url_tarball="${FZF_RELEASE_BASE}/${fzf_tag}/${fzf_artifact}"
  fzf_tmpdir="$(mktemp -d)"
  fzf_tarball="${fzf_tmpdir}/${fzf_artifact}"
  echo "Baixando fzf ${fzf_tag} (${fzf_artifact})..."
  download_file "$fzf_url_tarball" "$fzf_tarball" || {
    echo "Falha ao baixar fzf: ${fzf_url_tarball}" >&2
    exit 1
  }
  tar -xzf "$fzf_tarball" -C "$fzf_tmpdir"
  fzf_binary="$(find "$fzf_tmpdir" -maxdepth 2 -name fzf -type f | head -n1)"
  [ -n "$fzf_binary" ] || { echo "Binário fzf não encontrado no tarball." >&2; exit 1; }
  cp -f "$fzf_binary" "${INSTALL_DIR}/fzf"
  chmod +x "${INSTALL_DIR}/fzf"
  echo "fzf ${fzf_tag} instalado em ${INSTALL_DIR}/fzf"
}

install_yq_manual() {
  local yq_tag yq_os yq_arch yq_artifact yq_url yq_stage yq_fallback dest
  yq_fallback="${HOME}/.local/share/mb-cli/bin"
  yq_tag="$(get_yq_latest_tag)"
  [ -n "$yq_tag" ] || { echo "Não foi possível obter a última versão do yq." >&2; exit 1; }
  case "$OS" in
    linux)  yq_os="linux" ;;
    darwin) yq_os="darwin" ;;
    *)      echo "OS não mapeado para yq: $OS" >&2; exit 1 ;;
  esac
  case "$ARCH" in
    amd64) yq_arch="amd64" ;;
    arm64) yq_arch="arm64" ;;
    *)     echo "ARCH não mapeado para yq: $ARCH" >&2; exit 1 ;;
  esac
  yq_artifact="yq_${yq_os}_${yq_arch}"
  yq_url="${YQ_RELEASE_BASE}/${yq_tag}/${yq_artifact}"
  yq_stage="$(mktemp -d)"
  echo "Baixando yq ${yq_tag} (${yq_artifact})..."
  if ! download_file "$yq_url" "${yq_stage}/yq.bin"; then
    rm -rf "$yq_stage"
    echo "Falha ao baixar yq: ${yq_url}" >&2
    exit 1
  fi
  chmod +x "${yq_stage}/yq.bin"
  dest="${INSTALL_DIR}/yq"
  if cp -f "${yq_stage}/yq.bin" "$dest" 2>/dev/null && chmod +x "$dest" 2>/dev/null; then
    echo "yq ${yq_tag} instalado em ${dest}"
    rm -rf "$yq_stage"
    return 0
  fi
  mkdir -p "$yq_fallback"
  if cp -f "${yq_stage}/yq.bin" "${yq_fallback}/yq" && chmod +x "${yq_fallback}/yq"; then
    rm -rf "$yq_stage"
    echo "Aviso: não foi possível gravar yq em ${dest} (permissões ou ficheiro existente de outro dono)." >&2
    echo "yq ${yq_tag} foi instalado em ${yq_fallback}/yq (sem sudo). Adicione ao PATH ou remova o yq antigo em ${INSTALL_DIR}." >&2
    MB_YQ_EXTRA_PATH="$yq_fallback"
    return 0
  fi
  rm -rf "$yq_stage"
  echo "Falha ao instalar yq em ${dest} e em ${yq_fallback}." >&2
  exit 1
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

  # macOS + Homebrew: instala dependências via brew
  if [ "$OS" = "darwin" ] && command -v brew >/dev/null 2>&1; then
    echo "Instalando dependências via Homebrew (gum, glow, jq, fzf, yq)..."
    brew install gum glow jq fzf yq || echo "Aviso: brew install de dependências falhou; instale manualmente se necessário." >&2
    
    # mas (https://github.com/mas-cli/mas) — CLI da Mac App Store
    echo "Instalando mas via Homebrew (Mac App Store CLI)..."
    brew install mas || echo "Aviso: brew install mas falhou; continue manualmente se precisar do mas." >&2
    
    echo ""
    echo "MB CLI instalado em ${INSTALL_DIR}."
    echo "Dependências (gum, glow, jq, fzf, yq) instaladas via Homebrew."
  else
    # Linux ou macOS sem Homebrew: instala manualmente via GitHub Releases
    MB_YQ_EXTRA_PATH=""

    # Instalar gum (dependência do MB CLI) no mesmo INSTALL_DIR
    install_gum_manual
    
    # Instalar glow (dependência do MB CLI) no mesmo INSTALL_DIR
    install_glow_manual
    
    # Instalar jq (binário direto, sem tarball; sem checksum no release)
    install_jq_manual
    
    # Instalar fzf no mesmo INSTALL_DIR
    install_fzf_manual
    
    # Instalar yq (mikefarah/yq) no mesmo INSTALL_DIR
    install_yq_manual

    echo ""
    if [ -n "${MB_YQ_EXTRA_PATH:-}" ]; then
      echo "MB CLI, gum, glow, jq e fzf em ${INSTALL_DIR}; yq em ${MB_YQ_EXTRA_PATH}."
    else
      echo "MB CLI, gum, glow, jq, fzf e yq foram instalados em ${INSTALL_DIR}."
    fi
  fi
  if ! echo "$PATH" | grep -qF "${INSTALL_DIR}"; then
    # Escolhe o arquivo de config do shell; só adiciona se o path ainda não estiver lá
    if [ -n "${ZSH_VERSION:-}" ] || [ "$(basename "$SHELL" 2>/dev/null)" = "zsh" ]; then
      rcfile="${HOME}/.zshrc"
    elif [ "$(basename "$SHELL" 2>/dev/null)" = "bash" ]; then
      rcfile="${HOME}/.bashrc"
    else
      rcfile="${HOME}/.profile"
    fi
    if [ -f "$rcfile" ] && grep -q '\.local/bin' "$rcfile" 2>/dev/null; then
      echo "O arquivo $rcfile já parece conter ${INSTALL_DIR} no PATH. Abra um novo terminal ou rode: source $rcfile"
    else
      printf '\n# MB CLI (install.sh)\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$rcfile"
      echo "${INSTALL_DIR} foi adicionado ao PATH em $rcfile. Abra um novo terminal ou rode: source $rcfile"
    fi
  fi
  if [ -n "${MB_YQ_EXTRA_PATH:-}" ] && ! echo "$PATH" | grep -qF "${MB_YQ_EXTRA_PATH}"; then
    if [ -n "${ZSH_VERSION:-}" ] || [ "$(basename "$SHELL" 2>/dev/null)" = "zsh" ]; then
      rcfile="${HOME}/.zshrc"
    elif [ "$(basename "$SHELL" 2>/dev/null)" = "bash" ]; then
      rcfile="${HOME}/.bashrc"
    else
      rcfile="${HOME}/.profile"
    fi
    if [ -f "$rcfile" ] && grep -qF '.local/share/mb-cli/bin' "$rcfile" 2>/dev/null; then
      echo "O arquivo $rcfile já parece conter o diretório de yq (fallback) no PATH. Abra um novo terminal ou rode: source $rcfile"
    else
      printf '\n# MB CLI (install.sh) — yq (diretório fallback)\nexport PATH="%s:$PATH"\n' "$MB_YQ_EXTRA_PATH" >> "$rcfile"
      echo "${MB_YQ_EXTRA_PATH} foi adicionado ao PATH em $rcfile. Abra um novo terminal ou rode: source $rcfile"
    fi
  fi
  echo "Depois rode: mb plugins sync"
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
