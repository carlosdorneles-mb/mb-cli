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

## Ordem de precedência {#ordem-de-precedencia}

Da **menor** para a **maior** precedência:

1. **Variáveis do sistema** — O que já está em `os.Environ()` (incluindo o que você exportou no shell).
2. **`env.defaults`** — `~/.config/mb/env.defaults` (secrets **`.secrets`** no keyring; **`*.opsecrets`** com **`op://`** resolvidos via 1Password CLI quando a integração está disponível).
3. **Vault (`--env-vault`)** — Se você passar **`--env-vault <nome>`**, o arquivo `~/.config/mb/.env.<nome>` é mesclado por cima do `env.defaults` (secrets e **`op://`** desse vault da mesma forma; mesmas chaves sobrescrevem as do default).
4. **`mbcli.yaml` → `envs`** — Chave de topo **`envs`** no ficheiro **`mbcli.yaml`** do projeto. Deve ser um **mapa** YAML: chaves cujo valor é **escalar** entram no vault de projeto **`default`**; chaves cujo valor é um **mapa** definem vaults nomeados no próprio ficheiro (ex.: `staging:` com mais pares `NOME: valor` escalares). Os nomes dos vaults aninhados seguem as mesmas regras que `~/.config/mb/.env.<vault>`. **Não** há suporte a segredos neste ficheiro (sem `.secrets`, keyring nem `op://` em `mbcli.yaml` — só valores em claro). O caminho do ficheiro segue a mesma regra dos shell helpers: **`MBCLI_YAML_PATH`** (prioridade), senão **`${MBCLI_PROJECT_ROOT}/mbcli.yaml`** com raiz relativa ao cwd quando não for absoluto; se **`MBCLI_PROJECT_ROOT`** estiver vazio, usa-se o diretório atual. Se o ficheiro não existir ou **`envs`** estiver ausente/vazio, esta etapa é ignorada. YAML inválido ou **`envs`** que não seja um mapa faz o comando falhar com mensagem clara. Na mescla do ambiente: aplicam-se **sempre** as entradas do vault de projeto **`default`** (escalares na raiz de **`envs`**). Só quando **`--env-vault <nome>`** está definido é que, **além disso**, se mesclam as entradas de **`envs.<nome>`** por cima (sobrescrevendo chaves iguais vindas do default do projeto). Sem **`--env-vault`**, sub-mapas nomeados **não** entram no ambiente (alinha ao comportamento de não haver overlay de vault só pelo ficheiro).
5. **`.env` no diretório atual** — Se existir um ficheiro **`.env`** no **diretório de trabalho atual** (cwd) quando o comando corre, as variáveis dele são mescladas a seguir. Se o ficheiro não existir, esta etapa é ignorada. Erros de leitura (permissões, formato inválido) fazem o comando falhar com mensagem clara.
6. **`--env-file <path>`** — Mesclado em seguida e sobrescreve chaves anteriores em caso de conflito.
7. **`env_files` do manifest** (só **plugins**) — Arquivos `.env` declarados no `manifest.yaml` do plugin para o **vault efetivo**: sem **`--env-vault`**, entram só entradas com **`vault: default`** (ou `vault` omitido no YAML); com **`--env-vault test`**, entram só entradas com **`vault: test`**. Vários arquivos para o mesmo vault são aplicados **na ordem** do manifest (o último vence em chaves repetidas). Paths são relativos à pasta do plugin e não podem sair dela. O comando **`mb run`** não usa manifest de plugin, logo esta camada não se aplica.
8. **`--env KEY=VALUE`** — Maior precedência (pode ser repetido).

Ou seja: sem **`--env-vault`**, só entram as variáveis de `env.defaults` como base (além do sistema); com vault, o arquivo na config complementa/sobrescreve o default; em seguida entram as variáveis da chave **`envs`** em **`mbcli.yaml`** (default do projeto e, com **`--env-vault`**, o sub-mapa desse nome quando existir), depois o `./.env` do cwd (se existir), depois **`--env-file`**, depois os `env_files` do manifest (em comandos de plugin), e por fim **`--env`** tem a última palavra.

Exemplo de **`envs`** com default e vault nomeado:

```yaml
envs:
  API_BASE: https://api.example.test
  FEATURE_X: "1"
  staging:
    API_BASE: https://api.staging.example
```

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

Além de **`env.defaults`**, as entradas escalares na raiz de **`envs`** em **`mbcli.yaml`** são sempre mescladas. Com **`--env-vault <nome>`**, o ficheiro **`~/.config/mb/.env.<nome>`** e o sub-mapa homónimo em **`mbcli.yaml`** (quando existir) aplicam-se por cima.

## Como usar

### Defaults persistentes

Variáveis definidas com `mb envs` são usadas em toda execução de plugins e `mb run`:

```bash
mb envs set API_KEY=seu-valor
mb envs set DB_URL=postgres://prod --vault production
```

Resumo dos subcomandos (referência completa em [`mb envs`](../commands/envs.md)):

- **`mb envs list`** — Tabela com colunas **VAR** (`KEY=VALUE`), **VAULT** e **ARMAZENAMENTO**. Para ficheiros em **`~/.config/mb`**, **VAULT** é `default` ou o nome do ficheiro **`.env.<vault>`**. Para variáveis vindas de **`mbcli.yaml` → `envs`**, **VAULT** aparece como **`project`** (escalares na raiz) ou **`project/<nome>`** (sub-mapas). **ARMAZENAMENTO** inclui `local`, `keyring`, `1password` e **`projeto`** para o YAML do projeto. Use **`--show-secrets`** para ver valores reais.
- **`mb envs list --json`** ou **`-J`** — Emite um único objeto JSON (`{"CHAVE":"valor", ...}`), útil para scripts. Não pode ser usado junto com **`--text` / `-T`**.
- **`mb envs list --text`** ou **`-T`** — Emite só linhas **`CHAVE=valor`** (sem coluna de vault). Não pode ser usado junto com **`--json` / `-J`**.
- **`mb envs list --vault <valor>`** — Com **`--vault staging`** (exemplo), lista o ficheiro **`~/.config/mb/.env.staging`** e o overlay correspondente em **`mbcli.yaml`** (`envs` na raiz + sub-mapa **`staging`** se existir). Com **`--vault project`**, lista **apenas** as chaves escalares na raiz de **`envs`** no **`mbcli.yaml`**. Com **`--vault project/staging`**, lista **apenas** as chaves dentro do sub-mapa **`envs.staging`** (sem as da raiz) — **sem** ler **`env.defaults`** nem **`.env.*`**. O nome **`project`** e o prefixo **`project/`** estão **reservados** para estes vaults lógicos; **não** podem ser usados em **`mb envs set --vault`** nem em **`--env-vault`** para ficheiros em **`~/.config/mb`**.
- **`mb envs vaults`** — Tabela **VAULT** / **ARQUIVO** / **ENVS** (número de variáveis por vault). Inclui **`default`**, vaults em disco **`~/.config/mb/.env.<nome>`** (exceto **`.env.project`**, ignorado) e linhas **`project`** / **`project/<nome>`** quando **`mbcli.yaml`** tiver **`envs`**. **`--json` / `-J`**: `[{"vault","path","env_count"},...]`.
- **`mb envs set KEY=VALOR [KEY2=VALOR2 ...]`** — Grava no ficheiro padrão ou com **`--vault <nome>`** no **`.env.<nome>`**. O nome do vault segue as regras habituais **e** não pode ser **`project`** nem começar por **`project/`**. Com **`--secret`**, o valor vai ao **keyring** (lista **`.secrets`**). Com **`--secret-op`**, valor no **1Password** e referência em **`*.opsecrets`**. Sem flags, pode usar a variável **`MB_ENVS_SECRET_STORE`** (`plain` / `keyring` / `op`) no ambiente do processo ou nos ficheiros alvo em texto claro.
- **`mb envs unset KEY [KEY2 ...]`** — Remove uma ou mais chaves do vault padrão ou de **`--vault`**. Se não restar conteúdo nem segredos num **vault explícito**, apaga **`.env.<vault>`**, **`.secrets`** e **`.opsecrets`** associados; **`env.defaults`** nunca é apagado por ficar vazio.

### Vault na linha de comando

```bash
mb --env-vault staging tools meu-comando
```

### Arquivo de ambiente

```bash
mb --env-file .env.production tools meu-comando
```

O conteúdo do arquivo é mesclado ao ambiente antes de rodar o plugin ou o **`mb run`** (por cima do `./.env` do cwd e de **`mbcli.yaml` → `envs`**, se existirem; em plugins, ainda por baixo dos `env_files` do manifest).

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
- [Plugins](../technical-reference/plugins.md) — Como o ambiente é injetado no processo do plugin (merge no código, incluindo `mbcli.yaml`)
- [Criar um plugin](../plugin-authoring/create-a-plugin.md) — `env_files` e declaração de .env no manifest
