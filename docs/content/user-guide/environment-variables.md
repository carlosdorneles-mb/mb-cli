---
sidebar_position: 5
---

# Variáveis de ambiente

O MB CLI controla o ambiente em que os **plugins** e o comando **`mb run`** são executados. As variáveis são mescladas em uma ordem bem definida e só o **processo filho** recebe esse ambiente final — o CLI não altera o ambiente do seu shell.

## Gerenciando variáveis

Para definir, listar e remover variáveis persistentes, use **`mb envs`**. Veja a [referência completa do comando](../commands/envs.md).

Resumo rápido:

```bash
mb envs list                    # Lista todas
mb envs set KEY=VALOR           # Define no vault padrão
mb envs set KEY=VALOR --secret  # Guarda no keyring
mb envs unset KEY               # Remove
```

## Ordem de precedência

Da **menor** para a **maior** precedência:

| # | Camada | Descrição |
|---|--------|-----------|
| 1 | **Sistema** | Variáveis já existentes no shell (`os.Environ()`) |
| 2 | **`env.defaults`** | Ficheiro base em `~/.config/mb/env.defaults` (ou equivalente no macOS) |
| 3 | **Vault (`--env-vault`)** | Overlay de um vault nomeado (ex.: `~/.config/mb/.env.staging`) |
| 4 | **`.env` no cwd** | Ficheiro no diretório de trabalho atual, se existir |
| 5 | **`--env-file`** | Ficheiro explicitamente passado na linha de comando |
| 6 | **`env_files` do manifest** | Só plugins — arquivos declarados no `manifest.yaml` para o vault efetivo |
| 7 | **`--env KEY=VALUE`** | Overrides na linha de comando (maior precedência) |

Ou seja: sem `--env-vault`, só o `env.defaults` entra como base. Com vault, o overlay complementa/sobrescreve. Em seguida entra o `./.env` do cwd, depois `--env-file`, depois os `env_files` do plugin e por fim `--env` tem a última palavra.

## Secrets e 1Password

### Keyring do sistema (`--secret`)

Variáveis definidas com `mb envs set KEY=VALOR --secret` não são gravadas em ficheiro: o valor fica no **keyring** (macOS Keychain, Linux Secret Service). A chave é listada num ficheiro `.secrets` ao lado do env. Na listagem, essas chaves aparecem com valor `***`; use `--show-secrets` para ver o valor real.

### 1Password (`--secret-op`)

O MB pode guardar valores no **1Password** via [1Password CLI](https://developer.1password.com/docs/cli/). O **valor** fica num item no cofre; a **referência `op://`** é gravada em `*.opsecrets`. Requer `op` no PATH e sessão ativa.

```bash
mb envs set API_TOKEN=seu-valor --secret-op
mb envs set API_TOKEN=seu-valor --secret-op --vault staging
```

No vault padrão, `--secret-op` pede confirmação. Em CI, use `--yes`. Para detalhes, veja [Integração com 1Password](../commands/envs.md#envs-1password-secret-op).

## Vaults

Vaults são overlays de variáveis por contexto (ex.: staging, production):

```bash
mb envs set DB_URL=postgres://staging --vault staging
mb --env-vault staging tools meu-comando
```

Sem `--env-vault`, apenas o `env.defaults` entra. Com vault, o ficheiro `.env.<nome>` é mesclado por cima.

## Como usar

### Defaults persistentes

Variáveis definidas com `mb envs` são usadas em toda execução de plugins e `mb run`:

```bash
mb envs set API_KEY=seu-valor
mb envs set DB_URL=postgres://prod --vault production
```

### Vault na linha de comando

```bash
mb --env-vault staging tools meu-comando
```

### Arquivo de ambiente

```bash
mb --env-file .env.production tools meu-comando
```

### Variáveis na linha de comando

```bash
mb --env API_KEY=xyz --env AMBIENTE=prod tools meu-comando
```

### Comando arbitrário: `mb run`

Para executar qualquer programa com o mesmo ambiente mesclado:

```bash
mb run python script.py
mb run uv sync
mb run --env-vault staging -e FOO=bar python script.py
```

As flags globais (`-e` / `--env`, `--env-file`, `--env-vault`, `-v`, `-q`) podem ir antes ou depois de `run`, sempre antes do executável. Depois de `--`, nada é interpretado como flag do MB.

## Tema padrão do gum

O CLI injeta variáveis `GUM_*` para que scripts de plugins que usam [gum](https://github.com/charmbracelet/gum) herdem as cores do MB: cabeçalhos em laranja (`#FFA500`) e opção selecionada em verde (`#00A86B`).

Variáveis injetadas:

- **choose** — `GUM_CHOOSE_HEADER_FOREGROUND`, `GUM_CHOOSE_SELECTED_FOREGROUND`, `GUM_CHOOSE_CURSOR_FOREGROUND`
- **input** — `GUM_INPUT_PROMPT_FOREGROUND`, `GUM_INPUT_HEADER_FOREGROUND`, `GUM_INPUT_CURSOR_FOREGROUND`
- **confirm** — `GUM_CONFIRM_PROMPT_FOREGROUND`, `GUM_CONFIRM_SELECTED_FOREGROUND`, `GUM_CONFIRM_SELECTED_BACKGROUND`
- **spin** — `GUM_SPIN_TITLE_FOREGROUND`, `GUM_SPIN_SPINNER_FOREGROUND`

Esses defaults só são aplicados quando a chave ainda não existe no ambiente mesclado. Para usar suas próprias cores, defina as mesmas chaves em `env.defaults` ou com `--env`.

## Exemplo prático

Num script de plugin `run.sh`:

```bash
#!/bin/sh
echo "API_KEY está definida? ${API_KEY:-não}"
```

Se você definiu `API_KEY` com `mb envs set API_KEY=seu-valor` ou com `--env API_KEY=abc`, o plugin verá o valor ao ser executado.

## Ver também

- [`mb envs`](../commands/envs.md) — Referência completa do comando (subcomandos, flags, 1Password)
- [`mb run`](../commands/run.md) — Executar comandos arbitrários com ambiente mesclado
- [Plugins](../technical-reference/plugins.md) — Como o ambiente é injetado no processo do plugin
- [Criar um plugin](../plugin-authoring/create-a-plugin.md) — `env_files` e declaração de .env no manifest
