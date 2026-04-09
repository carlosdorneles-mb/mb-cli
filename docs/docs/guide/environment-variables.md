---
sidebar_position: 5
---

# Variáveis de ambiente

O MB CLI controla o ambiente em que os **plugins** e o comando **`mb run`** são executados. As variáveis são mescladas em uma ordem bem definida e só o **processo filho** (plugin ou comando passado a `mb run`) recebe esse ambiente final; o CLI em si não altera o ambiente do seu shell.

## Secrets no keyring

Variáveis definidas com **`mb envs set <KEY> <VALUE> --secret`** não são gravadas em ficheiro: o valor fica no **keyring do sistema** (macOS Keychain, Linux Secret Service, Windows Credential Manager). Ao listar com **`mb envs list`**, essas chaves aparecem com valor **`***`**; use **`mb envs list --show-secrets`** para ver o valor real (lido do keyring). Ao executar plugins ou **`mb run`**, o CLI resolve os secrets a partir do keyring e injecta o valor real no ambiente do processo.

Para guardar o valor no **1Password** e só manter uma referência **`op://`** no keyring, use **`--secret-op`** (não combina com **`--secret`**); detalhes na secção [Integração com 1Password](#envs-1password-secret-op).

**`mb envs unset <KEY>`** remove a variável do ficheiro (ou só a entrada na lista de secrets, para chaves só-secret) e, se for secret, também do keyring; com referência **`op://`**, remove também o campo correspondente no 1Password. Se a chave **não existir** nesse grupo (nem no ficheiro nem na lista **`.secrets`** do grupo), o comando termina com sucesso (**código 0**) e informa que não há variável com esse nome — **não** altera ficheiros nem o keyring.

## Integração com 1Password (--secret-op) {#envs-1password-secret-op}

O MB pode guardar valores sensíveis no **1Password** via [1Password CLI](https://developer.1password.com/docs/cli/) (`op` no **PATH**). O fluxo é distinto de **`--secret`**: o **valor** fica num item no cofre 1Password; no keyring fica apenas uma **referência** no formato **`op://`** (e a chave continua listada no ficheiro **`.secrets`** do grupo, como nos outros secrets).

### Requisitos

- **`op` instalado e sessão ativa** — inicie sessão com a CLI conforme a documentação da 1Password (`op signin`, etc.). Sem `op` disponível, o MB sugere instalar com **`mb tools 1password-cli`** ou seguir [Get started with 1Password CLI](https://developer.1password.com/docs/cli/).
- As flags **`--secret`** e **`--secret-op`** são **mutuamente exclusivas**.

### Como definir

```bash
mb envs set API_TOKEN "seu-valor" --secret-op
mb envs set API_TOKEN "seu-valor" --secret-op --group staging
```

O MB cria ou reutiliza um item do tipo **senha** no 1Password por grupo lógico, com título **`mb-cli env / default`** (grupo padrão) ou **`mb-cli env / <nome-do-grupo>`** (com **`--group`**), e grava o valor num campo reservado ao MB. No keyring guarda a referência **`op://...`**; o ficheiro **`env.defaults`** (ou **`.env.<grupo>`**) **não** contém o valor em claro.

### Listagem e execução

- Com **`mb envs list --show-secrets`**, referências **`op://`** são **resolvidas** com **`op read`** (é preciso sessão 1Password válida).
- Ao mesclar o ambiente para **plugins** ou **`mb run`**, valores **`op://`** no keyring são também resolvidos via integração com o `op`. Se a integração ou o `op` não estiverem disponíveis, o comando falha com mensagem clara em vez de injetar a referência literal.

### Remover

**`mb envs unset <KEY>`** (com o mesmo **`--group`**, se aplicável) remove a chave da lista **`.secrets`**, apaga a entrada no keyring e remove o campo no item 1Password quando o valor armazenado era uma referência **`op://`**.

## Ordem de precedência

Da **menor** para a **maior** precedência:

1. **Variáveis do sistema** — O que já está em `os.Environ()` (incluindo o que você exportou no shell).
2. **`env.defaults`** — `~/.config/mb/env.defaults` (e secrets do grupo default resolvidos do keyring; valores guardados como **`op://`** são lidos com o 1Password CLI quando a integração está disponível).
3. **Grupo (`--env-group`)** — Se você passar **`--env-group <nome>`**, o arquivo `~/.config/mb/.env.<nome>` é mesclado por cima do `env.defaults` (e os secrets desse grupo são resolvidos do keyring, incluindo **`op://`**; mesmas chaves do grupo sobrescrevem as do default).
4. **`.env` no diretório atual** — Se existir um ficheiro **`.env`** no **diretório de trabalho atual** (cwd) quando o comando corre, as variáveis dele são mescladas a seguir. Se o ficheiro não existir, esta etapa é ignorada. Erros de leitura (permissões, formato inválido) fazem o comando falhar com mensagem clara.
5. **`--env-file <path>`** — Mesclado em seguida e sobrescreve chaves anteriores em caso de conflito.
6. **`env_files` do manifest** (só **plugins**) — Arquivos `.env` declarados no `manifest.yaml` do plugin para o **grupo efetivo**: sem `--env-group`, entram só entradas com grupo `default` (ou com `group` omitido no YAML); com **`--env-group test`**, entram só entradas com `group: test`. Vários arquivos para o mesmo grupo são aplicados **na ordem** do manifest (o último vence em chaves repetidas). Paths são relativos à pasta do plugin e não podem sair dela. O comando **`mb run`** não usa manifest de plugin, logo esta camada não se aplica.
7. **`--env KEY=VALUE`** — Maior precedência (pode ser repetido).

Ou seja: sem `--env-group`, só entram as variáveis de `env.defaults` como base (além do sistema); com grupo, o arquivo do grupo na config complementa/sobrescreve o default; em seguida entra o `./.env` do cwd (se existir), depois **`--env-file`**, depois os `env_files` do manifest (em comandos de plugin), e por fim **`--env`** tem a última palavra.

## Tema padrão do gum

O CLI injeta variáveis de ambiente **GUM_\*** para que os scripts dos plugins que usam [gum](https://github.com/charmbracelet/gum) (choose, input, confirm, spin, etc.) herdem as cores do MB: cabeçalhos/títulos em laranja (`#FFA500`) e opção selecionada/cursor em verde (`#00A86B`). Entre outras, são injetadas variáveis para:

- **choose** — `GUM_CHOOSE_HEADER_FOREGROUND`, `GUM_CHOOSE_SELECTED_FOREGROUND`, `GUM_CHOOSE_CURSOR_FOREGROUND`
- **input** — `GUM_INPUT_PROMPT_FOREGROUND`, `GUM_INPUT_HEADER_FOREGROUND`, `GUM_INPUT_CURSOR_FOREGROUND`
- **confirm** — `GUM_CONFIRM_PROMPT_FOREGROUND` (título), `GUM_CONFIRM_SELECTED_FOREGROUND` / `GUM_CONFIRM_SELECTED_BACKGROUND`
- **spin** — `GUM_SPIN_TITLE_FOREGROUND`, `GUM_SPIN_SPINNER_FOREGROUND`

Esses defaults só são aplicados quando a chave ainda não existe no ambiente mesclado. Para usar suas próprias cores, defina as mesmas chaves em **`~/.config/mb/env.defaults`** (por exemplo com `mb envs set GUM_CHOOSE_HEADER_FOREGROUND 99`) ou na linha de comando com **`--env`**.

## Como usar

### Defaults persistentes: `mb envs`

Você pode definir variáveis que serão usadas em toda execução de plugins e de **`mb run`**, sem precisar passar `--env` toda vez:

- **`mb envs list`** — Tabela com colunas **VAR** (`KEY=VALUE`) e **GRUPO** (`default` para `env.defaults`, ou o nome do grupo para `~/.config/mb/.env.<grupo>`). Variáveis guardadas como secret mostram o valor como **`***`**; use **`--show-secrets`** para ver o valor real (keyring; referências **`op://`** são resolvidas com o 1Password CLI).
- **`mb envs list --json`** ou **`-J`** — Emite um único objeto JSON (`{"CHAVE":"valor", ...}`), útil para scripts. Não pode ser usado junto com **`--text` / `-T`**.
- **`mb envs list --text`** ou **`-T`** — Emite só linhas **`CHAVE=valor`** (sem coluna de grupo). Não pode ser usado junto com **`--json` / `-J`**.
- **`mb envs list --group <grupo>`** — Lista só as variáveis desse grupo (arquivo `.env.<grupo>`).
- **`mb envs set <KEY> <VALUE>`** — Grava a env no arquivo padrão de variáveis de ambiente. Com **`--group <grupo>`**, grava no arquivo referente ao grupo. O nome do grupo só pode conter letras, números, `.`, `_` e `-`. Com **`--secret`**, o valor é guardado no **keyring do sistema** (não em ficheiro). Com **`--secret-op`**, o valor é guardado no **1Password** e no keyring fica só a referência **`op://`** (exige **`op`** no PATH e sessão válida); não usar **`--secret`** em conjunto. Ao executar plugins ou **`mb run`**, os valores são resolvidos a partir do keyring e, no caso **`op://`**, via CLI do 1Password.
- **`mb envs unset <KEY>`** — Remove do arquivo padrão de variáveis de ambiente ou, com **`--group`**, só do arquivo do grupo. Se a variável estava guardada como secret (lista **`.secrets`** desse ficheiro), é também removida do keyring (e do 1Password, se aplicável). Se **`KEY`** não estiver definida nesse grupo, mostra que não existe variável com esse nome e **não** grava nada (comportamento idempotente; saída **0**).

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

O conteúdo do arquivo é mesclado ao ambiente antes de rodar o plugin ou o **`mb run`** (por cima do `./.env` do cwd, se existir; em plugins, ainda por baixo dos `env_files` do manifest).

### Comando arbitrário: `mb run`

Resumo na [Referência de comandos](../technical-reference/reference.md) (tabela **Comandos principais**).

Para executar qualquer programa com o mesmo ambiente mesclado (útil para scripts, `python`, `uv`, etc.):

```bash
mb run python script.py
mb run uv sync
```

As flags globais do `mb` (`--env`, `--env-file`, `--env-group`, etc.) vêm **antes** de `run`. Exemplo:

```bash
mb --env-file .env.local run uv sync
```

O subprocesso herda stdin, stdout e stderr do terminal. O **código de saída** do programa filho é propagado (não fica sempre `1` em caso de falha). Para ajuda do subcomando use **`mb help run`** (com `mb run --help`, o `--help` pode ser repassado ao programa executado).

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

Se você definiu `API_KEY` com `mb envs set API_KEY seu-valor` ou com `--env API_KEY=abc`, o plugin verá o valor ao ser executado.

Para detalhes de implementação (onde no código o merge é feito e como é passado ao processo do plugin), veja a [Referência técnica](../technical-reference/plugins.md).
