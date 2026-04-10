---
name: mb-envs
description: >-
  Covers MB CLI persistent envs (`mb envs`), merge into plugin/`mb run` processes,
  keyring secrets, 1Password (`--secret-op`, `op://`), and `*.opsecrets`. Use when changing or
  explaining `mb envs` list/set/unset/vaults, `env.defaults`, `.env.<vault>`, `.secrets`,
  `.opsecrets`, `SecretStore`, `OnePasswordEnv`, `BuildEnvFileValues`, global flags `--env` /
  `--env-file` / `--env-vault`, or docs under environment-variables / security.
---

# MB CLI — `mb envs` e ambiente

## Quando aplicar

- Implementar ou corrigir **`mb envs`** (list, vaults, set, unset) ou caminhos `env.defaults` / `.env.<vault>`.
- Alterar **merge de ambiente** para plugins ou **`mb run`**.
- Trabalhar com **`--secret`**, **`--secret-op`**, keyring, **`*.opsecrets`**, **`op://`**, ou cliente **`op`** (1Password).
- Atualizar **documentação** de variáveis de ambiente ou segurança relacionada a secrets.

## Comandos (superfície)

| Comando | Notas |
|--------|--------|
| `mb envs list` | Tabela / `--json` / `--text`; `--show-secrets`; `--vault` |
| `mb envs vaults` | Tabela VAULT / ARQUIVO / ENVS; vaults `project` e `project/*` via `mbcli.yaml`; `--json` com `env_count` |
| `mb envs set KEY=VAL [...]` | Plain; `--vault`; **`--secret`** (keyring); **`--secret-op`** (1Password + `.opsecrets`); **`MB_ENVS_SECRET_STORE`**; **`--yes`** com `--secret-op` no default |
| `mb envs unset KEY [KEY...]` | Várias chaves; `--vault`; remove keyring / 1Password / `.opsecrets` quando aplicável |

## Onde está o código

| Área | Caminho |
|------|---------|
| Casos de uso | `internal/usecase/envs/` |
| Cobra / UX | `internal/cli/envs/` |
| Merge ficheiros + secrets + `op://` | `internal/deps/envdefaults.go`, `execenv.go`; merge final `internal/shared/env/merge.go` |
| Lista de chaves secretas | `internal/deps/secretkeys.go` |
| Referências 1Password em ficheiro | `internal/deps/opsecrets.go` |
| 1Password CLI | `internal/infra/opcli/` |
| Portas | `internal/ports/onepassword.go`; `deps.Dependencies` inclui `SecretStore` e `OnePassword` |
| Wiring FX | `internal/module/deps/fx.go` |

Detalhe por ficheiro: [reference.md](reference.md).

## Regras de produto (resumo)

1. **Keyring**: chaves em **`*.secrets`**; valores no keyring por namespace (`default` ou nome do `--vault`).
2. **`--secret-op` (novo)**: valor no 1Password; referência **`op://`** em **`*.opsecrets`** (não no keyring).
3. **Merge**: ver `docs/docs/guide/environment-variables.md` — sistema → defaults → `--env-vault` → `./.env` cwd → `--env-file` → manifest `env_files` → `--env`.
4. **`unset`**: chave ausente se não está no mapa do ficheiro nem em `.secrets` nem em `.opsecrets`.

## Armadilhas conhecidas (1Password / `op`)

- Erro **`isn't an item`** no `op item get`: `isItemNotFound` em `client.go` tem de reconhecer **`isn't`** (contracção), não só `is not an item`.
- **`op item edit` stdin**: remover **`vault`** do JSON antes de editar (`marshalItemJSONForOPEdit` em `itemjson.go`) para evitar inconsistência Private vs Employee.
- **`sudo mb`**: root usa **PATH** diferente — `mb` instalado em `~/.local/bin` ou via mise pode não existir; usar caminho absoluto ou não usar sudo no `mb` quando desnecessário.
