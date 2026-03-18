---
sidebar_position: 5
---

# Variáveis de ambiente

O MB CLI controla o ambiente em que os plugins são executados. As variáveis são mescladas em uma ordem bem definida e só o processo do plugin recebe esse ambiente final; o CLI em si não altera o ambiente do seu shell.

## Ordem de precedência

Da **menor** para a **maior** precedência:

1. **Variáveis do sistema** — O que já está em `os.Environ()` (incluindo o que você exportou no shell).
2. **`env.defaults`** — `~/.config/mb/env.defaults`.
3. **Grupo (`--env-group`)** — Se você passar **`--env-group <nome>`**, o arquivo `~/.config/mb/.env.<nome>` é mesclado por cima do `env.defaults` (mesmas chaves do grupo sobrescrevem as do default).
4. **`--env-file <path>`** — Mesclado em seguida e sobrescreve chaves anteriores em caso de conflito.
5. **`--env KEY=VALUE`** — Maior precedência (pode ser repetido).

Ou seja: sem `--env-group`, só entram as variáveis de `env.defaults` como base (além do sistema); com grupo, o arquivo do grupo complementa/sobrescreve o default. **`--env`** continua com a última palavra.

## Tema padrão do gum

O CLI injeta variáveis de ambiente **GUM_\*** para que os scripts dos plugins que usam [gum](https://github.com/charmbracelet/gum) (choose, input, confirm, spin, etc.) herdem as cores do MB: cabeçalhos/títulos em laranja (`#FFA500`) e opção selecionada/cursor em verde (`#00A86B`). Entre outras, são injetadas variáveis para:

- **choose** — `GUM_CHOOSE_HEADER_FOREGROUND`, `GUM_CHOOSE_SELECTED_FOREGROUND`, `GUM_CHOOSE_CURSOR_FOREGROUND`
- **input** — `GUM_INPUT_PROMPT_FOREGROUND`, `GUM_INPUT_HEADER_FOREGROUND`, `GUM_INPUT_CURSOR_FOREGROUND`
- **confirm** — `GUM_CONFIRM_PROMPT_FOREGROUND` (título), `GUM_CONFIRM_SELECTED_FOREGROUND` / `GUM_CONFIRM_SELECTED_BACKGROUND`
- **spin** — `GUM_SPIN_TITLE_FOREGROUND`, `GUM_SPIN_SPINNER_FOREGROUND`

Esses defaults só são aplicados quando a chave ainda não existe no ambiente mesclado. Para usar suas próprias cores, defina as mesmas chaves em **`~/.config/mb/env.defaults`** (por exemplo com `mb self env set GUM_CHOOSE_HEADER_FOREGROUND 99`) ou na linha de comando com **`--env`**.

## Como usar

### Defaults persistentes: `mb self env`

Você pode definir variáveis que serão usadas em toda execução de plugins, sem precisar passar `--env` toda vez:

- **`mb self env list`** — Tabela com colunas **VAR** (`KEY=VALUE`) e **GRUPO** (`default` para `env.defaults`, ou o nome do grupo para `~/.config/mb/.env.<grupo>`).
- **`mb self env list --group <grupo>`** — Lista só as variáveis desse grupo (arquivo `.env.<grupo>`).
- **`mb self env set <KEY> <VALUE>`** — Grava a env no arquivo padrão de variáveis de ambiente. Com **`--group <grupo>`**, grava no arquivo referente ao grupo. O nome do grupo só pode conter letras, números, `.`, `_` e `-`.
- **`mb self env unset <KEY>`** — Remove do arquivo padrão de variáveis de ambiente ou, com **`--group`**, só do arquivo do grupo.

### Grupo na linha de comando: `--env-group`

Para uma execução usar o overlay de um grupo (por exemplo staging):

```bash
mb --env-group staging tools meu-comando
```

Carrega `env.defaults` e depois aplica `~/.config/mb/.env.staging` por cima. Sem `--env-group`, apenas o `env.defaults` entra nessa camada (como antes).

### Arquivo de ambiente: `--env-file`

Para usar um arquivo `.env` em um caminho específico (por exemplo, por projeto):

```bash
mb --env-file .env tools meu-comando
```

O conteúdo do arquivo é mesclado ao ambiente antes de rodar o plugin.

### Variáveis na linha de comando: `--env`

Para sobrescrever ou definir variáveis em uma única execução:

```bash
mb --env API_KEY=xyz --env AMBIENTE=prod tools meu-comando
```

## Exemplo prático

No seu plugin, você pode acessar variáveis injetadas normalmente. Por exemplo, em um script `run.sh`:

```bash
#!/bin/sh
echo "API_KEY está definida? ${API_KEY:-não}"
```

Se você definiu `API_KEY` com `mb self env set API_KEY seu-valor` ou com `--env API_KEY=abc`, o plugin verá o valor ao ser executado.

Para detalhes de implementação (onde no código o merge é feito e como é passado ao processo do plugin), veja a [Referência técnica](../technical-reference/plugins.md).
