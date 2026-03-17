#!/bin/bash

# GitHub Release Management helper: fetch versions, download and install releases.
# Compatible with Linux and macOS (POSIX sh).
#
# Usage:
#   . "$MB_HELPERS_PATH/github.sh"
#
# Public functions:
#   github_get_latest_version               - Get latest release tag from GitHub API
#   github_get_version_from_raw             - Get version from a raw JSON file in the repo
#   github_get_latest_version_with_fallback - Get latest version with API/raw fallback
#   github_detect_os_arch                   - Detect OS and architecture for downloads
#   github_build_download_url               - Build a release download URL from a pattern
#   github_download_release                 - Download a release file
#   github_verify_checksum                  - Verify SHA256/SHA512/MD5 checksum of a file
#   github_download_and_verify              - Download and verify a release
#   github_extract_tarball                  - Extract a tar.gz file
#   github_install_binary                   - Install a binary from an extracted archive
#   github_install_release                  - Full pipeline: download, verify and install

. "$MB_HELPERS_PATH/log.sh"

# Prints the latest release tag of a GitHub repository.
# Usage:
#   github_get_latest_version <owner/repo> [strip_prefix]
# Example:
#   version=$(github_get_latest_version "cli/cli" "true")
github_get_latest_version() {
    local repo="$1"
    local strip_prefix="${2:-false}"
    local api_url="https://api.github.com/repos/${repo}/releases/latest"
    local max_time="${GITHUB_API_MAX_TIME:-10}"
    local connect_timeout="${GITHUB_API_CONNECT_TIMEOUT:-5}"

    log debug "Consultando última versão de $repo..."

    local version=$(curl -s \
        --max-time "$max_time" \
        --connect-timeout "$connect_timeout" \
        "$api_url" 2> /dev/null |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        log error "Não foi possível obter a versão mais recente de $repo"
        log debug "Verifique sua conexão com a internet ou tente novamente"
        return 1
    fi

    # Remove prefixo 'v' se solicitado
    if [ "$strip_prefix" = "true" ]; then
        version="${version#v}"
    fi

    log debug "Última versão: $version"
    echo "$version"
}

# Prints a version field value from a raw JSON file in a GitHub repository.
# Usage:
#   github_get_version_from_raw <owner/repo> [branch] <file_path> [json_field]
# Example:
#   version=$(github_get_version_from_raw "user/repo" "main" "core/cli.json" "version")
github_get_version_from_raw() {
    local repo="$1"
    local branch="${2:-main}"
    local file_path="$3"
    local json_field="${4:-version}"
    local raw_url="https://raw.githubusercontent.com/${repo}/${branch}/${file_path}"
    local max_time="${GITHUB_API_MAX_TIME:-10}"
    local connect_timeout="${GITHUB_API_CONNECT_TIMEOUT:-5}"

    log debug "Obtendo versão de $file_path (branch: $branch)..."

    local temp_file="/tmp/github_raw_$$_${RANDOM}"

    if curl -sSL \
        --max-time "$max_time" \
        --connect-timeout "$connect_timeout" \
        "$raw_url" -o "$temp_file" 2> /dev/null; then

        if [[ -f "$temp_file" ]] && [[ -s "$temp_file" ]]; then
            local version=$(jq -r ".$json_field // empty" "$temp_file" 2> /dev/null)
            rm -f "$temp_file"

            if [[ -n "$version" ]]; then
                log debug "Versão obtida via raw content: $version"
                echo "$version"
                return 0
            fi
        fi
    fi

    rm -f "$temp_file"
    log debug "Falha ao obter versão do arquivo raw"
    return 1
}

# Prints the latest version with fallback: tries GitHub API first, then raw file content.
# Output format: "version|method" (e.g. "1.0.0|api" or "1.0.0|raw").
# Usage:
#   github_get_latest_version_with_fallback <owner/repo> [branch] [file_path] [json_field]
# Example:
#   result=$(github_get_latest_version_with_fallback "user/repo" "main" "core/cli.json" "version")
github_get_latest_version_with_fallback() {
    local repo="$1"
    local branch="${2:-main}"
    local file_path="${3:-core/cli.json}"
    local json_field="${4:-version}"

    log debug "Tentando obter última versão de $repo..."

    # Try GitHub API first (for releases)
    local version=$(github_get_latest_version "$repo" 2> /dev/null)
    if [ -n "$version" ]; then
        echo "$version|api"
        return 0
    fi

    log debug "API falhou, tentando via raw content..."

    # Fallback to raw content
    version=$(github_get_version_from_raw "$repo" "$branch" "$file_path" "$json_field")
    if [ $? -eq 0 ] && [ -n "$version" ]; then
        echo "$version|raw"
        return 0
    fi

    log debug "Todos os métodos falharam"
    return 1
}

# Detects the current OS and architecture, printing it as "os:arch" (e.g. "linux:x64").
# Usage:
#   github_detect_os_arch [format]
# Example:
#   os_arch=$(github_detect_os_arch "standard")
github_detect_os_arch() {
    local os_format="${1:-standard}" # standard, darwin-macos, linux-gnu

    # Detectar OS
    local os_name=""
    case "$(uname -s)" in
        Linux*)
            case "$os_format" in
                darwin-macos) os_name="linux" ;;
                linux-gnu) os_name="linux-gnu" ;;
                *) os_name="linux" ;;
            esac
            ;;
        Darwin*)
            case "$os_format" in
                darwin-macos) os_name="darwin" ;;
                *) os_name="macos" ;;
            esac
            ;;
        *)
            log error "Sistema operacional não suportado: $(uname -s)"
            return 1
            ;;
    esac

    # Detectar arquitetura
    local arch=$(uname -m)
    case "$arch" in
        x86_64) arch="x64" ;;
        aarch64 | arm64) arch="arm64" ;;
        armv7l) arch="armv7" ;;
        i686) arch="x86" ;;
        *)
            log error "Arquitetura não suportada: $arch"
            return 1
            ;;
    esac

    log debug "Sistema detectado: ${os_name}:${arch}"
    echo "${os_name}:${arch}"
}

# Builds a GitHub release download URL by substituting {version}, {os} and {arch} in a pattern.
# Usage:
#   github_build_download_url <owner/repo> <version> <os_name> <arch> <file_pattern>
# Example:
#   url=$(github_build_download_url "cli/cli" "v2.0.0" "linux" "x64" "gh_{version}_{os}_{arch}.tar.gz")
github_build_download_url() {
    local repo="$1"
    local version="$2"
    local os_name="$3"
    local arch="$4"
    local file_pattern="$5"

    # Remove 'v' prefix se presente na versão
    local version_clean="${version#v}"

    # Substitui placeholders
    local filename="${file_pattern//\{version\}/$version_clean}"
    filename="${filename//\{os\}/$os_name}"
    filename="${filename//\{arch\}/$arch}"

    local url="https://github.com/${repo}/releases/download/${version}/${filename}"

    log debug "URL de download: $url"
    echo "$url"
}

# Downloads a release file from a URL to a local path. Returns 0 on success, 1 on error.
# Usage:
#   github_download_release <url> <output_file> [description]
# Example:
#   github_download_release "https://github.com/.../file.tar.gz" "/tmp/file.tar.gz" "CLI binary"
github_download_release() {
    local download_url="$1"
    local output_file="$2"
    local description="${3:-arquivo}"

    log info "Baixando $description..."
    log debug "URL: $download_url"
    log debug "Destino: $output_file"

    # Criar diretório se não existir
    mkdir -p "$(dirname "$output_file")"

    # Download com barra de progresso
    if curl -L --progress-bar \
        --connect-timeout 30 \
        --max-time 300 \
        --retry 3 \
        --retry-delay 2 \
        "$download_url" -o "$output_file" 2>&1 |
        while IFS= read -r line; do
            # Redireciona progresso para stderr
            echo "$line" >&2
        done; then

        if [ ! -f "$output_file" ] || [ ! -s "$output_file" ]; then
            log error "Arquivo baixado está vazio ou não foi criado"
            rm -f "$output_file"
            return 1
        fi

        log debug "Download concluído: $(du -h "$output_file" | cut -f1)"
        return 0
    else
        local exit_code=$?
        log error "Falha ao baixar $description (código: $exit_code)"
        rm -f "$output_file"
        return 1
    fi
}

# Verifies the checksum of a file against a checksum file or a direct hash value.
# Supported algorithms: sha256 (default), sha512, md5.
# Usage:
#   github_verify_checksum <file> <checksum_file_or_hash> [algorithm]
# Example:
#   github_verify_checksum "/tmp/tool.tar.gz" "/tmp/tool.tar.gz.sha256"
github_verify_checksum() {
    local file="$1"
    local checksum_source="$2"
    local algorithm="${3:-sha256}"

    if [ ! -f "$file" ]; then
        log error "Arquivo não encontrado: $file"
        return 1
    fi

    # Determinar comando de hash
    local hash_cmd=""
    case "$algorithm" in
        sha256)
            if command -v sha256sum &> /dev/null; then
                hash_cmd="sha256sum"
            elif command -v shasum &> /dev/null; then
                hash_cmd="shasum -a 256"
            else
                log warn "Comando sha256sum não disponível, pulando verificação"
                return 0
            fi
            ;;
        sha512)
            if command -v sha512sum &> /dev/null; then
                hash_cmd="sha512sum"
            elif command -v shasum &> /dev/null; then
                hash_cmd="shasum -a 512"
            else
                log warn "Comando sha512sum não disponível, pulando verificação"
                return 0
            fi
            ;;
        md5)
            if command -v md5sum &> /dev/null; then
                hash_cmd="md5sum"
            elif command -v md5 &> /dev/null; then
                hash_cmd="md5 -r"
            else
                log warn "Comando md5sum não disponível, pulando verificação"
                return 0
            fi
            ;;
        *)
            log error "Algoritmo não suportado: $algorithm"
            return 1
            ;;
    esac

    # Calcular hash do arquivo
    local computed_hash=$($hash_cmd "$file" | awk '{print $1}')
    log debug "Hash calculado: $computed_hash"

    # Obter hash esperado
    local expected_hash=""
    if [ -f "$checksum_source" ]; then
        # É um arquivo de checksum
        log debug "Lendo checksum de arquivo: $checksum_source"
        local filename=$(basename "$file")

        # Tentar encontrar o hash do arquivo (com ou sem ./ no início)
        expected_hash=$(grep -i "$filename" "$checksum_source" | awk '{print $1}' | head -1)

        # Se não encontrou, tenta com ./ no início
        if [ -z "$expected_hash" ]; then
            expected_hash=$(grep -i "./$filename" "$checksum_source" | awk '{print $1}' | head -1)
        fi

        # Se ainda não encontrou, tenta pegar a primeira linha
        if [ -z "$expected_hash" ]; then
            expected_hash=$(head -1 "$checksum_source" | awk '{print $1}')
        fi
    else
        # É o hash direto
        expected_hash="$checksum_source"
    fi

    if [ -z "$expected_hash" ]; then
        log warn "Não foi possível obter hash esperado, pulando verificação"
        return 0
    fi

    log debug "Hash esperado: $expected_hash"

    # Comparar hashes (case insensitive)
    if [ "${computed_hash,,}" = "${expected_hash,,}" ]; then
        log infocess "✓ Verificação de integridade bem-sucedida"
        return 0
    else
        log error "✗ Verificação de integridade falhou!"
        log error "  Esperado: $expected_hash"
        log error "  Obtido:   $computed_hash"
        return 1
    fi
}

# Downloads a release file and verifies its checksum using the GitHub releases URL convention.
# Usage:
#   github_download_and_verify <owner/repo> <version> <download_url> <output_file> <checksum_filename> [algorithm]
# Example:
#   github_download_and_verify "cli/cli" "v2.0.0" "$url" "/tmp/gh.tar.gz" "gh_checksums.txt"
github_download_and_verify() {
    local repo="$1"
    local version="$2"
    local download_url="$3"
    local output_file="$4"
    local checksum_filename="$5"
    local algorithm="${6:-sha256}"

    # Construir URL do checksum
    local checksum_url=""
    if [ -n "$checksum_filename" ] && [ "$checksum_filename" != "none" ]; then
        checksum_url="https://github.com/${repo}/releases/download/${version}/${checksum_filename}"
    fi

    # Chamar a função interna com URLs completas
    github_download_and_verify_internal "$download_url" "$checksum_url" "$output_file" "$algorithm" "release"
}

# Downloads and verifies a file using direct URLs. Used internally by github_download_and_verify.
# Usage:
#   github_download_and_verify_internal <download_url> <checksum_url> <output_file> [algorithm] [description]
github_download_and_verify_internal() {
    local download_url="$1"
    local checksum_url="$2"
    local output_file="$3"
    local algorithm="${4:-sha256}"
    local description="${5:-release}"

    # Baixar arquivo principal
    if ! github_download_release "$download_url" "$output_file" "$description"; then
        return 1
    fi

    # Se não há URL de checksum, retorna sucesso
    if [ -z "$checksum_url" ] || [ "$checksum_url" = "none" ]; then
        log debug "Verificação de checksum não disponível"
        return 0
    fi

    # Baixar arquivo de checksum
    local checksum_file="${output_file}.${algorithm}"
    log info "Baixando checksum..."

    if ! github_download_release "$checksum_url" "$checksum_file" "checksum" 2> /dev/null; then
        log warn "Não foi possível baixar arquivo de checksum, pulando verificação"
        return 0
    fi

    # Verificar checksum
    if ! github_verify_checksum "$output_file" "$checksum_file" "$algorithm"; then
        rm -f "$output_file" "$checksum_file"
        return 1
    fi

    # Limpar arquivo de checksum
    rm -f "$checksum_file"
    return 0
}

# Extracts a tar.gz file and prints the extraction directory path.
# Usage:
#   github_extract_tarball <tar_file> [extract_dir]
# Example:
#   dir=$(github_extract_tarball "/tmp/tool.tar.gz")
github_extract_tarball() {
    local tar_file="$1"
    local extract_dir="${2:-/tmp/github-extract-$$}"

    if [ ! -f "$tar_file" ]; then
        log error "Arquivo não encontrado: $tar_file"
        return 1
    fi

    log info "Extraindo arquivo..."
    log debug "Origem: $tar_file"
    log debug "Destino: $extract_dir"

    mkdir -p "$extract_dir"

    if tar -xzf "$tar_file" -C "$extract_dir" 2>&1 |
        while read -r line; do log debug "tar: $line"; done; then

        log debug "Extração concluída"
        echo "$extract_dir"
        return 0
    else
        log error "Falha ao extrair arquivo"
        rm -rf "$extract_dir"
        return 1
    fi
}

# Finds and installs a binary from an extracted archive directory. Prints the installed path.
# Usage:
#   github_install_binary <extracted_dir> <binary_name> <install_dir>
# Example:
#   github_install_binary "/tmp/extract-123" "gh" "/usr/local/bin"
github_install_binary() {
    local extracted_dir="$1"
    local binary_name="$2"
    local install_dir="$3"

    # Procurar binário
    local binary_path=$(find "$extracted_dir" -type f -name "$binary_name" | head -1)

    if [ -z "$binary_path" ]; then
        log error "Binário '$binary_name' não encontrado no arquivo extraído"
        rm -rf "$extracted_dir"
        return 1
    fi

    log debug "Binário encontrado: $binary_path"

    # Criar diretório de instalação
    mkdir -p "$install_dir"

    # Instalar binário
    local target_path="$install_dir/$binary_name"
    mv "$binary_path" "$target_path"
    chmod +x "$target_path"

    # Limpar arquivos temporários
    rm -rf "$extracted_dir"

    log debug "Binário instalado em: $target_path"
    echo "$target_path"
}

# Full install pipeline: detects OS/arch, downloads, verifies and installs a GitHub release binary.
# Usage:
#   github_install_release <owner/repo> <version> <binary_name> <install_dir> <file_pattern> [checksum_pattern] [algorithm]
# Example:
#   github_install_release "cli/cli" "v2.0.0" "gh" "/usr/local/bin" "gh_{version}_{os}_{arch}.tar.gz"
github_install_release() {
    local repo="$1"
    local version="$2"
    local binary_name="$3"
    local install_dir="$4"
    local file_pattern="$5"
    local checksum_pattern="${6:-}"
    local algorithm="${7:-sha256}"

    # Detectar OS e arquitetura
    local os_arch
    os_arch=$(github_detect_os_arch "standard")
    [ $? -ne 0 ] && return 1

    local os_name="${os_arch%:*}"
    local arch="${os_arch#*:}"

    # Construir URLs
    local download_url=$(github_build_download_url "$repo" "$version" "$os_name" "$arch" "$file_pattern")
    local checksum_url=""

    if [ -n "$checksum_pattern" ]; then
        checksum_url=$(github_build_download_url "$repo" "$version" "$os_name" "$arch" "$checksum_pattern")
    fi

    # Arquivo temporário
    local temp_file="/tmp/${binary_name}-${version}.tar.gz"

    # Baixar e verificar
    if ! github_download_and_verify_internal "$download_url" "$checksum_url" "$temp_file" "$algorithm" "$binary_name"; then
        return 1
    fi

    # Extrair
    local extracted_dir=$(github_extract_tarball "$temp_file")
    if [ $? -ne 0 ]; then
        rm -f "$temp_file"
        return 1
    fi

    rm -f "$temp_file"

    # Instalar binário
    github_install_binary "$extracted_dir" "$binary_name" "$install_dir"
}
