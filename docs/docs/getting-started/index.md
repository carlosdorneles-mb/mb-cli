---
sidebar_position: 1
---

# Começar

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

Garanta que `~/.local/bin` está no seu `PATH`. Depois rode `mb plugins sync` para atualizar o cache de plugins e os helpers de shell.

Para **atualizar plugins, o binário do MB e o sistema**, use **`mb update`** sem flags. A atualização de **pacotes do SO** corre no plugin **`machine/update`** (shell); sem esse plugin instalado e **`mb plugins sync`**, a fase de sistema não faz nada útil. Use **`--only-plugins`** e/ou **`--only-cli`** para escolher fases; **`--only-tools`** é atalho para **`mb tools --update-all`** (mesma fase) e só aparece na ajuda quando o pacote **`tools`** está no cache com essa flag no manifest. **`--only-system`** delega em **`mb machine update`** (mesma fase) e só aparece na ajuda quando **`machine/update`** está no cache. Pode **combinar** várias flags **`--only-*`**. Para **só o binário do MB CLI** (release oficial), use **`mb update --only-cli`**. **`--check-only`** só junto de **`--only-cli`**. Só aplica a binários com versão embutida (GitHub Release); se compilaste localmente ou usaste `go install`, o comando indica que uses `install.sh`. Linux/macOS, amd64/arm64. No Linux, a fase APT pode pedir **`sudo`** interativo.

**`mb update --only-cli --check-only`** — mesma condição; saída **`2`** se houver release mais nova (útil em scripts; usar throttle para a API do GitHub).

**Remover o CLI:**

```bash
curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/uninstall.sh | bash
```

Ou remova manualmente: `rm -f ~/.local/bin/mb ~/.local/bin/gum ~/.local/bin/glow ~/.local/bin/jq ~/.local/bin/fzf` (se foram instalados pelo install.sh). Os dados (plugins, configuração) permanecem em `~/.config/mb` (Linux) ou `~/Library/Application Support/mb` (macOS) e não são apagados.

Próximo passo: [Desenvolvimento Local](./local-development.md) para compilar a partir do código ou [Criar um plugin](../plugin-authoring/create-a-plugin.md) para montar seu primeiro plugin.
