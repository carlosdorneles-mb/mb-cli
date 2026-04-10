---
sidebar_position: 5
---

# Variáveis de ambiente

O MB CLI controla o ambiente em que os **plugins** e o comando **`mb run`** são executados. As variáveis são mescladas em uma ordem bem definida e só o **processo filho** (plugin ou comando passado a `mb run`) recebe esse ambiente final; o CLI em si não altera o ambiente do seu shell.

## Secrets no keyring

Variáveis definidas com **`mb envs set KEY=VALOR --secret`** não são gravadas em ficheiro: o valor fica no **keyring do sistema** (macOS Keychain, Linux Secret Service, Windows Credential Manager). A chave é listada no ficheiro **`.secrets`** ao lado do `.env` / `env.defaults`. Ao listar com **`mb envs list`**, essas chaves aparecem com valor **`***`**; use **`mb envs list --show-secrets`** para ver o valor real (lido do keyring). Ao executar plugins ou **`mb run`**, o CLI resolve os secrets a partir do keyring e injecta o valor real no ambiente do processo.

Para guardar o valor no **1Password** com referência **`op://`** no ficheiro **`*.opsecrets`** (e não no keyring), use **`--secret-op`** (não combina com **`--secret`**); detalhes na secção [Integração com 1Password](#envs-1password-secret-op).

**`mb envs unset`** aceita várias chaves; remove do ficheiro, da lista **`.secrets`** / **`*.opsecrets`**, do keyring e do 1Password conforme o caso. Se a chave **não existir** nesse vault, o comando termina com sucesso (**código 0**) e informa — **não** altera ficheiros nem stores.

## Integração com 1Password (--secret-op) {#envs-1password-secret-op}

O MB pode guardar valores sensíveis no **1Password** via [1Password CLI](https://developer.1password.com/docs/cli/) (`op` no **PATH**). O fluxo é distinto de **`--secret`**: o **valor** fica num item no cofre 1Password; a **referência** **`op://`** é gravada no ficheiro **`*.opsecrets`** (dotenv `KEY=op://...`) ao lado do env — **não** no keyring. Entradas antigas que ainda tenham **`op://`** só no keyring continuam a ser resolvidas até migrar.

### Requisitos

- **`op` instalado e sessão ativa** — inicie sessão com a CLI conforme a documentação da 1Password (`op signin`, etc.). Sem `op` disponível, o MB sugere instalar com **`mb tools 1password-cli`** ou seguir [Get started with 1Password CLI](https://developer.1password.com/docs/cli/).
- As flags **`--secret`** e **`--secret-op`** são **mutuamente exclusivas**.

### Como definir

```bash
mb envs set API_TOKEN=seu-valor --secret-op
mb envs set API_TOKEN=seu-valor --secret-op --vault staging
mb envs set A=1 B=2 C=3 --secret
```

O MB cria ou reutiliza um item do tipo **senha** no 1Password por vault lógico, com título **`mb-cli env / default`** (vault padrão) ou **`mb-cli env / <nome-do-vault>`** (com **`--vault`**), e grava o valor num campo reservado ao MB. A referência **`op://...`** fica em **`env.defaults.opsecrets`** ou **`.env.<vault>.opsecrets`**; o ficheiro principal **não** contém o valor em claro.

No **vault padrão**, **`--secret-op`** pede confirmação (pode haver pedidos frequentes de desbloqueio do 1Password em qualquer comando `mb` que resolva esse ambiente). Em CI ou pipes, use **`--yes`** para confirmar sem prompt.

### Listagem e execução

- Com **`mb envs list --show-secrets`**, referências **`op://`** (ficheiro ou keyring legado) são **resolvidas** com **`op read`** (é preciso sessão 1Password válida).
- Ao mesclar o ambiente para **plugins** ou **`mb run`**, entradas em **`*.opsecrets`** e valores **`op://`** ainda no keyring são resolvidos via integração com o `op`. Se a integração ou o `op` não estiverem disponíveis, o comando falha com mensagem clara em vez de injetar a referência literal.

### Remover

**`mb envs unset`** (com o mesmo **`--vault`**, se aplicável) remove a chave dos ficheiros, de **`.secrets`**, de **`*.opsecrets`**, do keyring e do item 1Password quando aplicável.

## Ordem de precedência

Da **menor** para a **maior** precedência:

1. **Variáveis do sistema** — O que já está em `os.Environ()` (incluindo o que você exportou no shell).
2. **`env.defaults`** — `~/.config/mb/env.defaults` (secrets **`.secrets`** no keyring; **`*.opsecrets`** com **`op://`** resolvidos via 1Password CLI quando a integração está disponível).
3. **Vault (`--env-vault`)** — Se você passar **`--env-vault <nome>`**, o arquivo `~/.config/mb/.env.<nome>` é mesclado por cima do `env.defaults` (secrets e **`op://`** desse vault da mesma forma; mesmas chaves sobrescrevem as do default).
4. **`.env` no diretório atual** — Se existir um ficheiro **`.env`** no **diretório de trabalho atual** (cwd) quando o comando corre, as variáveis dele são mescladas a seguir. Se o ficheiro não existir, esta etapa é ignorada. Erros de leitura (permissões, formato inválido) fazem o comando falhar com mensagem clara.
5. **`--env-file <path>`** — Mesclado em seguida e sobrescreve chaves anteriores em caso de conflito.
6. **`env_files` do manifest** (só **plugins**) — Arquivos `.env` declarados no `manifest.yaml` do plugin para o **vault efetivo**: sem **`--env-vault`**, entram só entradas com **`vault: default`** (ou `vault` omitido no YAML); com **`--env-vault test`**, entram só entradas com **`vault: test`**. Vários arquivos para o mesmo vault são aplicados **na ordem** do manifest (o último vence em chaves repetidas). Paths são relativos à pasta do plugin e não podem sair dela. O comando **`mb run`** não usa manifest de plugin, logo esta camada não se aplica.
7. **`--env KEY=VALUE`** — Maior precedência (pode ser repetido).

Ou seja: sem **`--env-vault`**, só entram as variáveis de `env.defaults` como base (além do sistema); com vault, o arquivo na config complementa/sobrescreve o default; em seguida entra o `./.env` do cwd (se existir), depois **`--env-file`**, depois os `env_files` do manifest (em comandos de plugin), e por fim **`--env`** tem a última palavra.

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

- **`mb envs list`** — Tabela com colunas **VAR** (`KEY=VALUE`), **VAULT** (`default` para `env.defaults`, ou o nome do vault para `~/.config/mb/.env.<vault>`) e **ARMAZENAMENTO** (`local`; `keyring`; `1password` para segredos 1Password ou referências **`op://`** legadas no keyring). Use **`--show-secrets`** para ver valores reais.
- **`mb envs list --json`** ou **`-J`** — Emite um único objeto JSON (`{"CHAVE":"valor", ...}`), útil para scripts. Não pode ser usado junto com **`--text` / `-T`**.
- **`mb envs list --text`** ou **`-T`** — Emite só linhas **`CHAVE=valor`** (sem coluna de vault). Não pode ser usado junto com **`--json` / `-J`**.
- **`mb envs list --vault <nome>`** — Lista só as variáveis desse vault (ficheiro `.env.<nome>`).
- **`mb envs vaults`** — Tabela **VAULT** / **ARQUIVO**: o vault **`default`** aponta para **`env.defaults`**; cada **`~/.config/mb/.env.<vault>`** existente aparece como linha adicional. **`--json` / `-J`** emite `[{"vault":"...","path":"..."},...]`.
- **`mb envs set KEY=VALOR [KEY2=VALOR2 ...]`** — Grava no ficheiro padrão ou com **`--vault <nome>`** no **`.env.<nome>`**. O nome do vault só pode conter letras, números, `.`, `_` e `-`. Com **`--secret`**, o valor vai ao **keyring** (lista **`.secrets`**). Com **`--secret-op`**, valor no **1Password** e referência em **`*.opsecrets`**. Sem flags, pode usar a variável **`MB_ENVS_SECRET_STORE`** (`plain` / `keyring` / `op`) no ambiente do processo ou nos ficheiros alvo em texto claro.
- **`mb envs unset KEY [KEY2 ...]`** — Remove uma ou mais chaves do vault padrão ou de **`--vault`**. Se não restar conteúdo nem segredos num **vault explícito**, apaga **`.env.<vault>`**, **`.secrets`** e **`.opsecrets`** associados; **`env.defaults`** nunca é apagado por ficar vazio.

### Vault na linha de comando: `--env-vault`

Para uma execução usar o overlay de um vault (por exemplo staging):

```bash
mb --env-vault staging tools meu-comando
```

Carrega `env.defaults` e depois aplica `~/.config/mb/.env.staging` por cima. Sem **`--env-vault`**, apenas o `env.defaults` entra nessa camada (como antes).

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

As flags globais do `mb` (`-e` / `--env`, `--env-file`, `--env-vault`, `-v` / `--verbose`, `-q` / `--quiet`) podem ir **antes** de `run` **ou** **logo após** `run`, sempre **antes do nome do executável**. O que vier depois do primeiro argumento posicional é repassado ao programa filho (ex.: `mb run grep -r padrão .`). Exemplos:

```bash
mb --env-file .env.local run uv sync
mb run --env-vault staging -e FOO=bar python script.py
```

Depois de `--` no `mb run`, nada é interpretado como flag do MB (útil se o filho precisar de argumentos que começam por `-`).

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

Se você definiu `API_KEY` com `mb envs set API_KEY=seu-valor` ou com `--env API_KEY=abc`, o plugin verá o valor ao ser executado.

Para detalhes de implementação (onde no código o merge é feito e como é passado ao processo do plugin), veja a [Referência técnica](../technical-reference/plugins.md).

## Paths e grupos de vault

- **`mb envs path`** — Mostra o caminho completo do ficheiro de ambiente padrão (`env.defaults`). Útil para confirmar qual ficheiro está a ser lido ou para copiar/backup.
- **`mb envs groups`** — Lista os grupos de ambiente disponíveis como tabela (nome do vault + caminho do ficheiro). Equivalente a `mb envs vaults`. Use **`--json`** para saída em JSON.
