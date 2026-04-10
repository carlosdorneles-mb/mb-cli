---
sidebar_position: 1
---

# `mb envs`

Gerencia variáveis de ambiente globais injetadas em plugins e no comando `mb run`.

Para entender ordem de precedência, vaults de projeto e quando usar cada abordagem, veja [Variáveis de ambiente](../user-guide/environment-variables.md).

## Subcomandos

### `mb envs list`

Lista variáveis de um vault.

```bash
mb envs list
mb envs list --vault staging
mb envs list --show-secrets
mb envs list --vault project        # só a raiz de envs no mbcli.yaml
mb envs list --vault project/staging # só o sub-mapa envs.staging
```

**Output:** tabela com colunas **VAR** (`KEY=VALUE`), **VAULT** e **ARMAZENAMENTO** (`local`, `keyring`, `1password`, `projeto` para `mbcli.yaml`).

| Flag | Descrição |
| ---- | --------- |
| `--vault <nome>` | Vault em disco `.env.<nome>` no diretório de configuração (ex.: `~/.config/mb/.env.staging` no Linux) com overlay do `mbcli.yaml`. Use `project` para a raiz de `envs` no YAML, ou `project/<nome>` para apenas o sub-mapa (sem misturar a raiz) |
| `--show-secrets` | Mostra valores reais em vez de `***` |
| `--json` / `-J` | Emite `{"CHAVE":"valor", ...}` |
| `--text` / `-T` | Emite `CHAVE=valor` por linha (sem coluna de vault) |

> `--json` e `--text` são mutuamente exclusivos.

### `mb envs set`

Define ou atualiza variáveis.

```bash
mb envs set API_KEY=xyz
mb envs set A=1 B=2 C=3 --vault staging
mb envs set DB_PASS=segredo --secret
mb envs set TOKEN=abc --secret-op
mb envs set A=1 B=2 C=3 --secret      # múltiplas chaves de uma vez
```

| Flag | Descrição |
|---|---|
| `--vault <nome>` | Grava em `.env.<nome>` em vez do padrão. Não aceita `project` nem prefixo `project/` (reservados) |
| `--secret` | Guarda no keyring do sistema (lista `.secrets`) |
| `--secret-op` | Guarda no 1Password (referência em `*.opsecrets`, requer `op`) |
| `--yes` | Confirma sem prompt para `--secret-op` (CI) |

> `--secret` e `--secret-op` são mutuamente exclusivos. Sem flags de store, a variável `MB_ENVS_SECRET_STORE` (`plain` / `keyring` / `op`) no ambiente define o destino.

### `mb envs unset`

Remove uma ou mais chaves.

```bash
mb envs unset API_KEY
mb envs unset A B C --vault staging
```

Se a chave não existir, exit code 0. Se não restar conteúdo num vault explícito, os ficheiros `.env.<vault>`, `.secrets` e `.opsecrets` são apagados. `env.defaults` nunca é removido.

### `mb envs vaults`

Lista vaults disponíveis, caminho e número de variáveis.

```bash
mb envs vaults
mb envs vaults --json
```

**Output:** tabela **VAULT** / **ARQUIVO** / **ENVS**. Inclui `default`, vaults em disco `.env.<nome>` no diretório de configuração (ex.: `~/.config/mb/.env.staging` no Linux), e linhas `project` / `project/<nome>` quando `mbcli.yaml` tem `envs`. O ficheiro `.env.project` é ignorado (nome reservado).

**JSON:** `[{"vault","path","env_count"},...]`

## Integração com 1Password (`--secret-op`) {#envs-1password-secret-op}

O MB guarda valores no **1Password** via [1Password CLI](https://developer.1password.com/docs/cli/) (`op` no PATH). O **valor** fica no cofre; a **referência `op://`** é gravada em `*.opsecrets` — **não** no keyring.

### Requisitos

- `op` instalado e sessão ativa (`op signin`)
- `--secret` e `--secret-op` são mutuamente exclusivos

### Comportamento

O MB cria ou reutiliza um item do tipo **senha** com título `mb-cli env / default` ou `mb-cli env / <vault>` por vault lógico. A referência `op://` fica em `env.defaults.opsecrets` ou `.env.<vault>.opsecrets`.

Na listagem com `--show-secrets`, referências `op://` são resolvidas com `op read`. Se a sessão 1Password não estiver ativa, o comando falha com mensagem clara.

`mb envs unset` (com o mesmo `--vault`) remove a chave dos ficheiros, `.secrets`, `*.opsecrets`, keyring e item 1Password.

## Ver também

- [Variáveis de ambiente](../user-guide/environment-variables.md) — Precedência, vaults de projeto, quando usar o quê
- [`mb run`](../commands/run.md) — Executar programas com ambiente mesclado
- [Comandos de plugins](../user-guide/plugin-commands.md) — Como plugins herdam o ambiente
