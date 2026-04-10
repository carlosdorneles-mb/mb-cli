---
sidebar_position: 1
---

# `mb envs`

Gerencia variáveis de ambiente globais que são injetadas em plugins e no comando `mb run`.

## Subcomandos

### `mb envs list`

Lista variáveis do vault padrão ou de um vault específico.

```bash
mb envs list
mb envs list --vault staging
mb envs list --show-secrets
```

**Output:** tabela com colunas **VAR** (`KEY=VALUE`), **VAULT** e **ARMAZENAMENTO** (`local`, `keyring`, `1password`).

| Flag | Descrição |
| ---- | --------- |
| `--vault <nome>` | Lista só variáveis do vault `.env.<nome>` |
| `--show-secrets` | Mostra valores reais em vez de `***` |
| `--json` / `-J` | Emite `{"CHAVE":"valor", ...}` |
| `--text` / `-T` | Emite `CHAVE=valor` por linha (sem vault) |

> `--json` e `--text` são mutuamente exclusivos.

### `mb envs set`

Define ou atualiza variáveis no vault padrão ou num vault específico.

```bash
mb envs set API_KEY=xyz
mb envs set A=1 B=2 C=3 --vault staging
mb envs set DB_PASS=segredo --secret
mb envs set TOKEN=abc --secret-op
```

| Flag | Descrição |
|---|---|
| `--vault <nome>` | Grava em `.env.<nome>` em vez do padrão |
| `--secret` | Guarda no keyring do sistema |
| `--secret-op` | Guarda no 1Password (requer `op` no PATH) |
| `--yes` | Confirma sem prompt para `--secret-op` (útil em CI) |

> `--secret` e `--secret-op` são mutuamente exclusivos. O nome do vault aceita letras, números, `.`, `_` e `-`.

### `mb envs unset`

Remove uma ou mais chaves do vault.

```bash
mb envs unset API_KEY
mb envs unset A B C --vault staging
```

Se a chave não existir, o comando termina com sucesso (código 0). Se não restar conteúdo num vault explícito, o ficheiro `.env.<vault>` e associados (`.secrets`, `.opsecrets`) são apagados. O `env.defaults` nunca é apagado.

### `mb envs vaults`

Lista vaults disponíveis e o caminho do ficheiro de cada um.

```bash
mb envs vaults
mb envs vaults --json
```

**Output:** tabela **VAULT** / **ARQUIVO**. O vault `default` aponta para `env.defaults`.

## Ordem de precedência

Quando um plugin ou `mb run` executa, as variáveis são mescladas nesta ordem (da menor para a maior precedência):

1. Variáveis do sistema (`os.Environ()`)
2. `env.defaults` (`~/.config/mb/env.defaults`)
3. Vault (`~/.config/mb/.env.<nome>`) com `--env-vault`
4. `.env` no diretório atual (cwd)
5. `--env-file <path>`
6. `env_files` do manifest (só plugins)
7. `--env KEY=VALUE` (maior precedência)

Para mais detalhes, veja [Variáveis de ambiente](../user-guide/environment-variables.md).
