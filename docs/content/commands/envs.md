---
sidebar_position: 1
---

# `mb envs`

Gerencia variáveis de ambiente globais que são injetadas em plugins e no comando `mb run`.

Para entender como as variáveis são mescladas e a ordem de precedência, veja [Variáveis de ambiente](../user-guide/environment-variables.md).

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

## Integração com 1Password (`--secret-op`) {#envs-1password-secret-op}

O MB pode guardar valores sensíveis no **1Password** via [1Password CLI](https://developer.1password.com/docs/cli/) (`op` no **PATH**). O fluxo é distinto de `--secret`: o **valor** fica num item no cofre 1Password; a **referência** `op://` é gravada no ficheiro `*.opsecrets` — **não** no keyring.

### Requisitos

- **`op` instalado e sessão ativa** — inicie sessão com a CLI conforme a documentação da 1Password (`op signin`, etc.). Sem `op` disponível, o MB sugere instalar com `mb tools 1password-cli`.
- As flags `--secret` e `--secret-op` são **mutuamente exclusivas**.

### Como funciona

O MB cria ou reutiliza um item do tipo **senha** no 1Password por vault lógico, com título `mb-cli env / default` (vault padrão) ou `mb-cli env / <nome-do-vault>` (com `--vault`), e grava o valor num campo reservado ao MB. A referência `op://...` fica em `env.defaults.opsecrets` ou `.env.<vault>.opsecrets`.

### Listagem e execução

- Com `mb envs list --show-secrets`, referências `op://` são **resolvidas** com `op read` (é preciso sessão 1Password válida).
- Ao mesclar o ambiente para **plugins** ou `mb run`, entradas em `*.opsecrets` e valores `op://` ainda no keyring são resolvidos via `op`. Se a integração não estiver disponível, o comando falha com mensagem clara.

### Remover

`mb envs unset` (com o mesmo `--vault`, se aplicável) remove a chave dos ficheiros, de `.secrets`, de `*.opsecrets`, do keyring e do item 1Password.

## Ver também

- [Variáveis de ambiente](../user-guide/environment-variables.md) — Ordem de precedência, conceito e uso prático
- [`mb run`](../commands/run.md) — Executar comandos com ambiente mesclado
- [Comandos de plugins](../user-guide/plugin-commands.md) — Como plugins herdam o ambiente
