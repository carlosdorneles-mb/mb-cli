# MB CLI

CLI em Go para orquestrar plugins com UX em laranja, descoberta dinâmica via cache SQLite e injeção segura de variáveis de ambiente.

## O que o MB faz

- **Comandos dinâmicos**: plugins viram comandos `mb <categoria> <comando>` automaticamente.
- **Cache SQLite**: após `mb self sync`, o CLI não precisa escanear o disco a cada execução.
- **Um manifesto por plugin**: cada plugin declara nome, categoria e, opcionalmente, subcategoria em `manifest.yaml` (ex.: `mb infra ci deploy`).
- **Ambiente controlado**: variáveis são mescladas (sistema → arquivo .env → `--env`) e injetadas só no processo do plugin.
- **Help e erros estilizados**: [Fang](https://github.com/charmbracelet/fang) estiliza o `--help` e as mensagens de erro com o tema padrão.

## Pré-requisitos

- Go 1.22+
- O **install.sh** instala o [gum](https://github.com/charmbracelet/gum) e o [glow](https://github.com/charmbracelet/glow) em `~/.local/bin` junto com o MB (sem root). O glow é necessário para a flag `--readme` dos plugins (help em Markdown).

## Instalação (usuários)

Para instalar o binário do MB sem compilar (Linux e macOS, amd64/arm64), use o script de instalação. Ele instala o MB CLI, o **gum** e o **glow** (dependências) em `~/.local/bin` (não requer sudo). O download é validado com o arquivo `checksums.txt` do release.

**Instalar (versão mais recente ou especificada):**

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash
# Ou com versão específica:
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash -s -- --version 0.0.5
```

Certifique-se de que `~/.local/bin` está no seu `PATH`. Depois rode `mb self sync` para preparar o cache e os helpers de shell.

**Remover o CLI:**

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/uninstall.sh | bash
```

Ou remova manualmente: `rm -f ~/.local/bin/mb ~/.local/bin/gum ~/.local/bin/glow` (se foram instalados pelo install.sh). Os dados do CLI (plugins, config) ficam em `~/.config/mb` (Linux) ou `~/Library/Application Support/mb` (macOS) e não são apagados na desinstalação.

## Build e instalação (desenvolvedores)

```bash
make build          # binário em bin/mb
make install        # instala em $GOPATH/bin
```

A versão exibida em `mb --version` vem do build: em desenvolvimento (`go run .` ou `make build`) aparece `dev`; em releases oficiais, a versão é injetada a partir da tag Git (ex.: `v1.0.0`).

## Versionamento e releases

O MB segue [Semantic Versioning](https://semver.org/) (SemVer). A versão é definida por **tags Git** (ex.: `v1.0.0`). Os releases são gerados com [GoReleaser](https://goreleaser.com/): ao dar push numa tag `v*`, o [workflow de release](.github/workflows/release.yml) roda no GitHub Actions, gera binários para **Linux** e **macOS** (amd64 e arm64) e publica um [GitHub Release](https://docs.github.com/en/repositories/releasing-projects-on-github) com os artefatos e o checksum.

**Como gerar uma nova versão**

1. Decidir a versão (ex.: `v1.0.0`).
2. Criar e enviar a tag: `git tag v1.0.0 && git push origin v1.0.0`.
3. O workflow de release dispara automaticamente e publica o release com os binários. O comando `mb --version` nos binários gerados exibirá essa versão.

## Como funciona a gestão de plugins

1. **Diretório de plugins**  
   O MB usa apenas um diretório, derivado de `os.UserConfigDir()`:
   - **Linux**: `~/.config/mb/plugins`
   - **macOS**: `~/Library/Application Support/mb/plugins`

2. **Descoberta**  
   O comando `mb self sync` percorre esse diretório em busca de arquivos `manifest.yaml`. Cada manifesto define um plugin (nome, categoria, tipo, script ou binário).

3. **Cache SQLite**  
   Os plugins encontrados são gravados em `~/.config/mb/cache.db` (ou equivalente no macOS). Esse banco é consultado na inicialização do CLI para montar a árvore de comandos. Por isso é importante rodar `mb self sync` depois de adicionar ou alterar plugins.

4. **Comandos no terminal**  
   O MB agrupa plugins por **categoria**. Cada categoria vira um subcomando, e cada plugin vira um subcomando dessa categoria:
   - Plugin `name: deploy`, `category: infra` → `mb infra deploy`
   - Plugin `name: lint`, `category: dev` → `mb dev lint`

5. **Execução**  
   Ao rodar `mb infra deploy`, o MB localiza o executável/script no cache, monta o ambiente (merge de env), e executa o processo. Você pode usar scripts shell (`type: sh`) ou binários (`type: bin`).

## Criar um plugin (passo a passo)

### 1. Criar o diretório do plugin

No diretório de plugins do MB (ex.: `~/.config/mb/plugins`):

```bash
# Linux
mkdir -p ~/.config/mb/plugins/meu-plugin

# macOS
mkdir -p ~/Library/Application\ Support/mb/plugins/meu-plugin
```

### 2. Criar o manifesto `manifest.yaml`

Crie `manifest.yaml` dentro da pasta do plugin:

```yaml
name: meu-plugin      # nome do comando (ex.: mb tools meu-plugin)
category: tools       # categoria = subcomando pai
type: sh              # sh = script shell; bin = executável
entrypoint: run.sh    # arquivo a executar (relativo à pasta do plugin)
readme: README.md     # opcional: usado pelo mb ... --help (glow)
```

Campos obrigatórios: `name`, `category`, `type`, `entrypoint`.  
`readme` é opcional; se existir, o `--help` do comando pode exibir o README via glow.

### 3. Criar o script ou binário

Para `type: sh`, crie o script referido em `entrypoint` (ex.: `run.sh`):

```bash
#!/bin/sh
echo "Plugin rodando!"
echo "Variável injetada: API_KEY=${API_KEY:-não definida}"
```

Torne o script executável:

```bash
chmod +x ~/.config/mb/plugins/meu-plugin/run.sh
```

Para `type: bin`, use um executável compilado (Go, Rust, etc.) e indique-o em `entrypoint`.

### 4. (Opcional) README para ajuda

Se você declarou `readme: README.md`, crie esse arquivo na mesma pasta. Ao rodar `mb tools meu-plugin --help`, o MB pode usar o glow para renderizar esse Markdown.

### 5. Registrar no cache e rodar

```bash
mb self sync
mb plugins list                 # lista plugins (incluindo meu-plugin)
mb tools meu-plugin             # executa o plugin
mb --env API_KEY=xyz tools meu-plugin   # com variável injetada
```

## Variáveis de ambiente

Antes de executar um plugin, o MB mescla:

1. Variáveis do sistema (`os.Environ()`)
2. Arquivo de defaults: `<UserConfigDir>/mb/env.defaults` (e opcionalmente `--env-file`)
3. Variáveis da linha de comando: `--env KEY=VALUE`

A **maior precedência** é a de `--env`. Assim você pode passar segredos só para o processo do plugin, sem deixá-los no histórico do shell.

Comandos para gerenciar defaults: `mb self env list`, `mb self env set KEY [VALUE]`, `mb self env unset KEY`.

## Comandos principais

| Comando | Descrição |
|--------|-----------|
| `mb self sync` | Escaneia o diretório de plugins e atualiza o cache SQLite |
| `mb plugins list` | Lista plugins instalados (name, command, version, url) |
| `mb self env list` | Lista variáveis padrão |
| `mb self env set KEY [VALUE]` | Define variável padrão |
| `mb self env unset KEY` | Remove variável padrão |
| `mb <categoria> <comando> [args...]` | Executa o plugin correspondente |

Flags globais: `--verbose`, `--quiet`, `--env-file <path>`, `--env KEY=VALUE`.

## Executar localmente

Para rodar o CLI sem instalar (comandos e exemplos completos em **[docs/running-locally.md](docs/running-locally.md)**):

```bash
make run-local              # go run . (ajuda: make run-local --help)
make run-local self sync    # argumentos podem ser passados direto
make run self sync          # build + ./bin/mb (idem: make run [args...] ou make run ARGS="...")
```

Para registrar os plugins de exemplo no seu config (sem copiar arquivos): **`make install-examples`**; em seguida rode `make run self sync`.

## Testar o CLI

```bash
make test       # testes unitários
make build && ./bin/mb self sync && ./bin/mb plugins list
```

Para testar sem alterar seu config real, use um diretório temporário:

```bash
export XDG_CONFIG_HOME=/tmp/mb-test   # Linux
mkdir -p "$XDG_CONFIG_HOME/mb/plugins/hello"
# ... criar manifest.yaml e run.sh ...
./bin/mb self sync
./bin/mb plugins list
./bin/mb <categoria> hello
```
