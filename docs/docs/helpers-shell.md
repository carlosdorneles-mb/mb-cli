---
sidebar_position: 7
---

# Helpers de shell

O MB CLI injeta no ambiente dos plugins a variável **`MB_HELPERS_PATH`**, que aponta para o **diretório** dos helpers de shell (`~/.config/mb/lib/shell`).

## Como carregar

No início do script do plugin (por exemplo em `run.sh`), importe o que precisar:

- **Todos os helpers:** `. "$MB_HELPERS_PATH/all.sh"`
- **Só o helper de log:** `. "$MB_HELPERS_PATH/log.sh"`
- **Só o helper de memória:** `. "$MB_HELPERS_PATH/memory.sh"`
- **Só o helper de string:** `. "$MB_HELPERS_PATH/string.sh"`
- **Só o helper de Kubernetes:** `. "$MB_HELPERS_PATH/kubernetes.sh"`
- **Só o helper de OS:** `. "$MB_HELPERS_PATH/os.sh"`
- **Só o helper de Snap:** `. "$MB_HELPERS_PATH/snap.sh"`
- **Só o helper de Homebrew:** `. "$MB_HELPERS_PATH/homebrew.sh"`
- **Só o helper de Flatpak:** `. "$MB_HELPERS_PATH/flatpak.sh"`
- **Só o helper de GitHub:** `. "$MB_HELPERS_PATH/github.sh"`

Exemplo:

```sh
#!/bin/sh
. "$MB_HELPERS_PATH/all.sh"

# A partir daqui você pode usar os helpers
log info "Olá!"
```

O diretório e os arquivos são criados ou atualizados quando você executa **`mb self sync`** (ou ao adicionar/atualizar plugins, que disparam o sync). Se os helpers ainda não existirem, execute `mb self sync` antes de usá-los nos seus plugins. Ao atualizar o CLI para uma versão que altere os helpers, o próximo `mb self sync` atualiza os arquivos em `lib/shell` automaticamente (o CLI compara um checksum do conteúdo embutido com o arquivo `.checksum` nesse diretório).

## Helpers disponíveis

### log

Log que respeita `MB_QUIET` e `MB_VERBOSE` (flags `-q` e `-v` do CLI). Usa `gum log -l` por baixo.

**Uso:** `log <level> <mensagem...>`

**Níveis:** `none`, `debug`, `info`, `warn`, `error`, `fatal`

**Comportamento:**

- **`MB_QUIET=1`** — Só exibe mensagens com nível `error` e `fatal`.
- **`MB_VERBOSE=1`** — Exibe todos os níveis, incluindo `debug`.
- **Caso contrário** — Exibe `info`, `warn`, `error`, `fatal`; o nível `debug` é omitido.

Exemplos:

```sh
log info "Processando..."
log debug "Detalhe interno: $var"
log warn "Aviso"
log error "Algo falhou"
```

### memory

Helper de memória simples (chave/valor) para scripts de plugin.

Ele salva dados em arquivos no diretório temporário do sistema (`${TMPDIR:-/tmp}/mb/memory`) usando a estrutura `namespace/key`. Isso permite reaproveitar respostas curtas do usuário em execuções futuras do mesmo plugin.

Como funciona:

- Cada valor fica em um arquivo: `${TMPDIR:-/tmp}/mb/memory/<namespace>/<key>`.
- O valor é sobrescrito quando você chama `mem_set` novamente para a mesma chave.
- A escrita é feita com arquivo temporário + `mv` (atômica) para reduzir risco de arquivo parcial.
- `namespace` e `key` aceitam somente letras, números, `.`, `_` e `-`.
- Por estar em `tmp`, o conteúdo pode ser removido pelo sistema (reboot/limpeza automática).

**Uso:**

- `mem_set <namespace> <key> <valor...>`
- `mem_get <namespace> <key> [default]`
- `mem_has <namespace> <key>`
- `mem_unset <namespace> <key>`
- `mem_clear_ns <namespace>`

**Comandos disponíveis:**

- `mem_set`: cria ou atualiza um valor.
	Ex.: `mem_set tools.deploy cluster prod`
- `mem_get`: lê um valor; se não existir, retorna o `default` (ou vazio).
	Ex.: `cluster="$(mem_get tools.deploy cluster dev)"`
- `mem_has`: verifica se a chave existe (ideal para `if`).
	Ex.: `if mem_has tools.deploy cluster; then ... fi`
- `mem_unset`: remove uma chave específica.
	Ex.: `mem_unset tools.deploy cluster`
- `mem_clear_ns`: remove todas as chaves de um namespace.
	Ex.: `mem_clear_ns tools.deploy`

Retornos:

- `0`: sucesso.
- `1`: ausência de chave em `mem_has` ou falha de I/O.
- `2`: `namespace`/`key` inválidos.

Exemplo:

```sh
. "$MB_HELPERS_PATH/memory.sh"

if ! mem_has "tools.deploy" "cluster"; then
	cluster="$(gum input --placeholder "Nome do cluster")"
	mem_set "tools.deploy" "cluster" "$cluster"
fi

cluster="$(mem_get "tools.deploy" "cluster" "dev")"
log info "Usando cluster: $cluster"
```

Observações:

- Esses dados ficam em `tmp` e podem ser removidos pelo sistema (por exemplo, em reboot ou limpeza automática).
- `namespace` e `key` aceitam somente letras, números, `.`, `_` e `-`.

### string

Helper de utilitários para manipulação de texto em scripts shell. Cobre substituição, conversão de case, trim, testes de conteúdo, manipulação de arrays CSV e conversão de booleano.

**Funções disponíveis:**

- `str_replace <input> <search> <replace>` — substitui todas as ocorrências de `search` por `replace` em `input` e imprime o resultado.
- `str_to_upper <texto>` — imprime o texto convertido para maiúsculas.
- `str_to_lower <texto>` — imprime o texto convertido para minúsculas.
- `str_trim <texto>` — imprime o texto sem espaços no início e no fim.
- `str_contains <texto> <substring>` — retorna `0` se `texto` contém `substring`, `1` caso contrário.
- `str_starts_with <texto> <prefixo>` — retorna `0` se `texto` começa com `prefixo`, `1` caso contrário.
- `str_parse_comma_separated <nome_array>` — percorre o array referenciado e divide cada elemento que contenha vírgula em elementos separados (modifica o array in-place).
- `str_join_to_comma_separated <nome_array>` — junta todos os elementos do array em um único elemento separado por vírgula (modifica o array in-place).
- `str_to_bool <valor>` — retorna `0` para valores verdadeiros (`true`, `1`, `on`, `yes`) e `1` para os demais.

Exemplo:

```sh
. "$MB_HELPERS_PATH/string.sh"

# Substituição e conversão
tag=$(str_to_lower "$(str_trim "  My-App  ")")
log info "Tag: $tag"  # my-app

# Testes condicionais
if str_starts_with "$tag" "my"; then
  log info "Tag começa com 'my'"
fi

# Booleano a partir de variável de ambiente
if str_to_bool "${DRY_RUN:-false}"; then
  log warn "Dry-run ativo, nenhuma alteração será feita"
fi
```

### kubernetes

Helper para operações básicas com `kubectl`: verificar se está instalado, checar existência de namespace e inspecionar o contexto ativo. Carrega `log.sh` automaticamente ao ser importado.

> **Requisito:** `kubectl` precisa estar instalado e configurado no `PATH`. Caso contrário, as funções logam um erro e, se `exit_on_error` for passado, encerram o script com `exit 1`.

**Funções disponíveis:**

- `kb_check_installed [exit_on_error]` — verifica se `kubectl` está disponível no `PATH`. Retorna `0` se encontrado, `1` se não. Com `exit_on_error`, encerra o script se não estiver instalado.
- `kb_check_namespace_exists <namespace> [exit_on_error]` — verifica se o namespace existe no cluster do contexto atual. Retorna `0` se existir, `1` se não. Com `exit_on_error`, encerra o script se não existir.
- `kb_get_current_context` — imprime o nome do contexto kubectl ativo (`kubectl config current-context`).
- `kb_print_current_context` — imprime o contexto atual no console com uma mensagem legível.

Exemplo:

```sh
. "$MB_HELPERS_PATH/kubernetes.sh"

# Garante que kubectl existe e que o namespace alvo também
kb_check_installed "exit_on_error"
kb_check_namespace_exists "production" "exit_on_error"

# Informa o contexto em uso antes de aplicar mudanças
kb_print_current_context
kubectl apply -f manifests/
```

### os

Helper para detecção de sistema operacional e distribuição Linux em scripts shell. Permite adaptar comportamentos de instalação e configuração ao ambiente do usuário.

**Funções disponíveis:**

- `get_simple_os` — imprime `linux`, `mac` ou `unknown`.
- `is_mac` — retorna `0` se estiver no macOS, `1` caso contrário.
- `is_linux` — retorna `0` se estiver no Linux, `1` caso contrário.
- `is_linux_debian` — retorna `0` em distros Debian-based (Ubuntu, Mint, Pop, etc.).
- `is_linux_redhat` — retorna `0` em distros RedHat-based (Fedora, RHEL, CentOS, Rocky, etc.).
- `is_linux_arch` — retorna `0` em distros Arch-based (Manjaro, EndeavourOS, etc.).
- `get_debian_pkg_manager` — imprime `apt-get` ou `apt`.
- `get_redhat_pkg_manager` — imprime `dnf` ou `yum`.
- `get_arch_pkg_manager` — imprime `pacman` ou `unknown`.
- `get_distro_id` — imprime o `$ID` de `/etc/os-release` (ex.: `ubuntu`, `fedora`, `arch`).

Exemplo:

```sh
. "$MB_HELPERS_PATH/os.sh"

os=$(get_simple_os)
log info "Sistema operacional: $os"

if is_linux_debian; then
  pkg=$(get_debian_pkg_manager)
  sudo "$pkg" install -y curl
elif is_linux_redhat; then
  pkg=$(get_redhat_pkg_manager)
  sudo "$pkg" install -y curl
elif is_mac; then
  brew install curl
fi
```

### snap

Helper para instalar, atualizar, remover e consultar aplicações via Snap Store. Compatível com sistemas Linux onde o `snapd` está disponível. Carrega `log.sh` automaticamente ao ser importado.

> **Requisito:** `snap` precisa estar instalado no sistema. As funções verificam sua disponibilidade e logam um erro se não encontrado.

**Funções disponíveis:**

- `snap_is_available` — retorna `0` se o `snap` está instalado.
- `snap_refresh_metadata` — atualiza os metadados do Snap Store (não crítico; retorna `0` sempre).
- `snap_is_installed <app_name>` — retorna `0` se a aplicação está instalada.
- `snap_get_installed_version <app_name>` — imprime a versão instalada ou `unknown`.
- `snap_get_latest_version <app_name>` — imprime a versão mais recente disponível ou `unknown`.
- `snap_install <app_name> [friendly_name] [channel] [classic]` — instala a aplicação; `classic` aceita `true`/`false` (padrão: `false`).
- `snap_update <app_name> [friendly_name] [channel]` — atualiza a aplicação.
- `snap_uninstall <app_name> [friendly_name]` — remove a aplicação.
- `snap_info <app_name>` — imprime informações detalhadas do pacote.
- `snap_list_installed` — lista todos os pacotes Snap instalados.

Exemplo:

```sh
. "$MB_HELPERS_PATH/snap.sh"

if ! snap_is_installed "podman-desktop"; then
  snap_install "podman-desktop" "Podman Desktop"
else
  snap_update "podman-desktop" "Podman Desktop"
fi
```

### homebrew

Helper para instalar, atualizar, remover e consultar casks e fórmulas via Homebrew no macOS. Carrega `log.sh` automaticamente ao ser importado.

> **Requisito:** `brew` precisa estar instalado. Para casks, as funções usam `brew install --cask`; para fórmulas, `brew install`.

**Funções de cask:**

- `homebrew_is_available` — retorna `0` se o `brew` está instalado.
- `homebrew_update_metadata` — executa `brew update` para atualizar os formulae.
- `homebrew_is_installed <cask_name>` — retorna `0` se o cask está instalado.
- `homebrew_get_installed_version <cask_name>` — imprime a versão instalada ou `unknown`.
- `homebrew_get_latest_version <cask_name>` — imprime a versão mais recente disponível ou `unknown`.
- `homebrew_install <cask_name> [friendly_name]` — instala o cask.
- `homebrew_update <cask_name> [friendly_name]` — atualiza o cask.
- `homebrew_uninstall <cask_name> [friendly_name]` — remove o cask (com `--zap`).

**Funções de fórmula:**

- `homebrew_is_installed_formula <formula_name>` — retorna `0` se a fórmula está instalada.
- `homebrew_get_installed_version_formula <formula_name>` — imprime a versão instalada ou `unknown`.
- `homebrew_get_latest_version_formula <formula_name>` — imprime a versão mais recente ou `unknown`.
- `homebrew_install_formula <formula_name> [friendly_name]` — instala a fórmula.
- `homebrew_update_formula <formula_name> [friendly_name]` — atualiza a fórmula.
- `homebrew_uninstall_formula <formula_name> [friendly_name]` — remove a fórmula.
- `homebrew_link_formula <formula_name> [force]` — cria os links simbólicos para os binários da fórmula.

Exemplo:

```sh
. "$MB_HELPERS_PATH/homebrew.sh"

homebrew_install "visual-studio-code" "VS Code"
homebrew_install_formula "libpq" "PostgreSQL client"
homebrew_link_formula "libpq" "true"
```

### flatpak

Helper para instalar, atualizar, remover e consultar aplicações via Flatpak a partir do Flathub. Compatível com sistemas Linux onde o `flatpak` está disponível. Carrega `log.sh` automaticamente ao ser importado.

> **Requisito:** `flatpak` precisa estar instalado. A função `flatpak_ensure_flathub` configure o repositório Flathub automaticamente se não estiver presente.

**Funções disponíveis:**

- `flatpak_is_available` — retorna `0` se o `flatpak` está instalado.
- `flatpak_ensure_flathub` — garante que o repositório Flathub está configurado (nível `--user`).
- `flatpak_update_metadata` — atualiza os metadados do Flathub (não crítico; retorna `0` sempre).
- `flatpak_is_installed <app_id>` — retorna `0` se a aplicação está instalada.
- `flatpak_get_installed_version <app_id>` — imprime a versão instalada ou `unknown`.
- `flatpak_get_latest_version <app_id>` — imprime a versão mais recente via Flathub API ou `unknown`.
- `flatpak_install <app_id> [friendly_name]` — instala a aplicação de Flathub.
- `flatpak_update <app_id> [friendly_name]` — atualiza a aplicação.
- `flatpak_uninstall <app_id> [friendly_name]` — remove a aplicação e seus dados.

Exemplo:

```sh
. "$MB_HELPERS_PATH/flatpak.sh"

flatpak_ensure_flathub
if ! flatpak_is_installed "io.podman_desktop.PodmanDesktop"; then
  flatpak_install "io.podman_desktop.PodmanDesktop" "Podman Desktop"
else
  flatpak_update "io.podman_desktop.PodmanDesktop" "Podman Desktop"
fi
```

### github

Helper para buscar versões, baixar e instalar releases do GitHub. Compatível com Linux e macOS. Carrega `log.sh` automaticamente ao ser importado.

> **Dependências:** `curl` (obrigatório), `jq` (opcional, usado em `github_get_version_from_raw`), `tar` (para `github_extract_tarball`). Variáveis opcionais: `GITHUB_API_MAX_TIME` (padrão: `10`) e `GITHUB_API_CONNECT_TIMEOUT` (padrão: `5`).

**Funções disponíveis:**

- `github_get_latest_version <owner/repo> [strip_prefix]` — imprime o tag da última release. Com `strip_prefix=true`, remove o prefixo `v`.
- `github_get_version_from_raw <owner/repo> [branch] <file_path> [json_field]` — imprime um campo de versão de um arquivo JSON raw no repositório.
- `github_get_latest_version_with_fallback <owner/repo> [branch] [file_path] [json_field]` — tenta a API primeiro; em fallback, usa o raw. Imprime `versão|método` (ex.: `1.0.0|api`).
- `github_detect_os_arch [format]` — imprime `os:arch` do sistema atual (ex.: `linux:x64`, `macos:arm64`).
- `github_build_download_url <owner/repo> <version> <os> <arch> <file_pattern>` — monta a URL substituindo `{version}`, `{os}` e `{arch}` no padrão.
- `github_download_release <url> <output_file> [description]` — baixa um arquivo de release.
- `github_verify_checksum <file> <checksum_file_or_hash> [algorithm]` — verifica checksum (`sha256`, `sha512` ou `md5`).
- `github_download_and_verify <owner/repo> <version> <url> <output_file> <checksum_filename> [algorithm]` — baixa e verifica usando a convenção de releases do GitHub.
- `github_extract_tarball <tar_file> [extract_dir]` — extrai um `.tar.gz` e imprime o diretório extraído.
- `github_install_binary <extracted_dir> <binary_name> <install_dir>` — localiza e instala o binário do arquivo extraído.
- `github_install_release <owner/repo> <version> <binary_name> <install_dir> <file_pattern> [checksum_pattern] [algorithm]` — pipeline completo: detecta OS/arch, baixa, verifica e instala.

Exemplo:

```sh
. "$MB_HELPERS_PATH/github.sh"

version=$(github_get_latest_version "cli/cli" "true")
os_arch=$(github_detect_os_arch)
os_name="${os_arch%:*}"
arch="${os_arch#*:}"

url=$(github_build_download_url "cli/cli" "v${version}" "$os_name" "$arch" "gh_{version}_{os}_{arch}.tar.gz")
github_download_and_verify "cli/cli" "v${version}" "$url" "/tmp/gh.tar.gz" "gh_${version}_checksums.txt"
dir=$(github_extract_tarball "/tmp/gh.tar.gz")
github_install_binary "$dir" "gh" "/usr/local/bin"
```
