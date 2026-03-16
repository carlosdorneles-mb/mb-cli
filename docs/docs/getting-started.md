---
sidebar_position: 1
---

# Começar

## Pré-requisitos

- Go 1.22+ (apenas se for compilar a partir do código)
- O **script de instalação** instala o [gum](https://github.com/charmbracelet/gum), o [glow](https://github.com/charmbracelet/glow), o [jq](https://github.com/jqlang/jq) e o [fzf](https://github.com/junegunn/fzf) em `~/.local/bin` junto com o MB (sem root). O glow é necessário para a flag `--readme` dos plugins (help em Markdown).

## Instalação do CLI

A forma recomendada é usar o **script de instalação**, que baixa o binário do MB CLI, o **gum**, o **glow**, o **jq** e o **fzf** (dependências) do [GitHub Releases](https://github.com/carlosdorneles-mb/mb-cli/releases) (e dos repositórios charmbracelet/gum, charmbracelet/glow, jqlang/jq e junegunn/fzf), valida os downloads com `checksums.txt` quando disponível e instala em **`~/.local/bin`** (sem precisar de sudo).

**Instalar:**

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash
```

Para uma versão específica:

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash -s -- --version 0.0.5
```

Garanta que `~/.local/bin` está no seu `PATH`. Depois rode `mb self sync` para atualizar o cache de plugins e os helpers de shell.

**Remover o CLI:**

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/uninstall.sh | bash
```

Ou remova manualmente: `rm -f ~/.local/bin/mb ~/.local/bin/gum ~/.local/bin/glow ~/.local/bin/jq ~/.local/bin/fzf` (se foram instalados pelo install.sh). Os dados (plugins, configuração) permanecem em `~/.config/mb` (Linux) ou `~/Library/Application Support/mb` (macOS) e não são apagados.

## Para desenvolvedores

### Build e instalação

Se você for alterar o código ou contribuir:

```bash
make build          # binário em bin/mb
make install        # instala em $GOPATH/bin
```

### Executar localmente

Para rodar o CLI sem instalar (a partir do código-fonte):

```bash
make run-local              # go run . (ajuda: make run-local --help)
make run-local self sync    # argumentos podem ser passados direto: make run-local [args...]
make run self sync          # build + ./bin/mb; idem: make run [args...] ou make run ARGS="..."
```

Para usar os plugins de exemplo do repositório: **`make install-examples`** (registra cada plugin com `mb plugins add`, sem copiar); depois **`make run self sync`** (ou `mb self sync`).

Próximo passo: [Criar um plugin](creating-plugins) para montar seu primeiro plugin e rodá-lo com o MB.
