---
sidebar_position: 5
---

# `mb completion`

Gera e gerencia scripts de autocompletar (TAB) para o MB CLI.

## Subcomandos

### `mb completion install`

Detecta o shell (variável `SHELL`) ou usa `--shell`, gera o script de autocompletar e grava um bloco idempotente no ficheiro de perfil (`~/.bashrc`, `~/.zshrc`, `fish/config.fish` ou perfil PowerShell).

```bash
mb completion install
mb completion install --shell zsh
mb completion install --dry-run      # pré-visualiza sem gravar
mb completion install --yes          # sem confirmação (útil em CI)
```

| Flag | Descrição |
|---|---|
| `--shell <nome>` | Força o shell (`bash`, `zsh`, `fish`, `powershell`) |
| `--rc-file <path>` | Ficheiro de perfil alternativo |
| `--dry-run` | Mostra o que seria gravado sem alterar nada |
| `--yes` | Confirma sem prompt (ambientes não interativos) |

> Em ambientes não interativos, `--yes` ou `--dry-run` é obrigatório.

### `mb completion uninstall`

Remove o bloco de autocompletar delimitado por marcadores `mb-cli` do ficheiro de perfil. Não altera nada se o ficheiro ou o bloco não existir.

```bash
mb completion uninstall
mb completion uninstall --dry-run
mb completion uninstall --yes
```

### Gerar scripts individualmente

Se quiser o script para um shell específico (por exemplo para distribuir ou incluir no seu próprio perfil):

```bash
mb completion bash
mb completion zsh
mb completion fish
mb completion powershell
```

Cada subcomando aceita `--no-descriptions` para desativar as descrições no autocompletar.

## Como funciona

O autocompletar inclui todos os comandos estáticos (`envs`, `plugins`, `run`, `update`, etc.) **e** os comandos dinâmicos montados a partir dos plugins instalados. Após cada `mb plugins sync`, o cache é atualizado e o autocompletar reflete os novos comandos.

Se instalar/remover plugins, pode regenerar o autocompletar com:

```bash
mb completion install --yes
```

Ou confiar que o sync do plugins já atualiza o autocompletar automaticamente.
