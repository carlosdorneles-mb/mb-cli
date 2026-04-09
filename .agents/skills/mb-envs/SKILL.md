---
name: mb-envs
description: >-
  Covers MB CLI persistent envs (`mb envs`), merge into plugin/`mb run` processes,
  keyring secrets, and 1Password (`--secret-op`, `op://`). Use when changing or
  explaining `mb envs` list/set/unset, `env.defaults`, `.env.<group>`, `.secrets`,
  `SecretStore`, `OnePasswordEnv`, `BuildEnvFileValues`, global flags `--env` /
  `--env-file` / `--env-group`, or docs under environment-variables / security.
---

# MB CLI — `mb envs` e ambiente

## Quando aplicar

- Implementar ou corrigir **`mb envs`** (list, groups, set, unset) ou caminhos `env.defaults` / `.env.<grupo>`.
- Alterar **merge de ambiente** para plugins ou **`mb run`**.
- Trabalhar com **`--secret`**, **`--secret-op`**, keyring, **`op://`**, ou cliente **`op`** (1Password).
- Atualizar **documentação** de variáveis de ambiente ou segurança relacionada a secrets.

## Comandos (superfície)

| Comando | Notas |
|--------|--------|
| `mb envs list` | Tabela / `--json` / `--text`; `--show-secrets`; `--group` |
| `mb envs groups` | Tabela GRUPO / ARQUIVO; alias `group`; `--json` / `-J` |
| `mb envs set KEY VAL` | Ficheiro plain; `--group`; **`--secret`** (valor no keyring); **`--secret-op`** (valor no 1Password, `op://` no keyring) — **mutuamente exclusivos** |
| `mb envs unset KEY` | `(removed bool)`; se chave inexistente no grupo → mensagem informativa, **não** grava; remove keyring e campo 1Password se `op://` |

## Onde está o código

| Área | Caminho |
|------|---------|
| Casos de uso | `internal/app/envs/` (`set.go`, `unset.go`, `list.go`, `groups.go`, `paths.go`) |
| Cobra / UX | `internal/cli/envs/` (`set.go`, `unset.go`, `list.go`, `groups.go`, `env.go`, `path.go`) |
| Merge ficheiros + secrets + `op://` | `internal/deps/envdefaults.go`, `execenv.go`; merge final `internal/shared/env/merge.go` |
| Lista de chaves secretas | `internal/deps/secretkeys.go` (ficheiro `path.secrets`) |
| 1Password CLI | `internal/infra/opcli/` (`client.go`, `itemjson.go`) |
| Portas | `internal/ports/onepassword.go`; `deps.Dependencies` inclui `SecretStore` e `OnePassword` |
| Wiring FX | `internal/module/deps/fx.go` (instancia `opcli`) |

Detalhe por ficheiro: [reference.md](reference.md).

## Regras de produto (resumo)

1. **Secrets só em ficheiro**: chaves listadas em **`*.secrets`** (mesmo basename que o `.env` / `env.defaults`); valores no **keyring** por grupo lógico (`default` ou nome do `--group`).
2. **`--secret-op`**: `PutSecret` grava no 1Password e devolve `op://`; keyring guarda a referência; **`op`** tem de existir e estar autenticado para set/list/merge.
3. **Merge** (ordem completa): ver `docs/docs/guide/environment-variables.md` — sistema → defaults → grupo → `./.env` cwd → `--env-file` → manifest `env_files` (só plugins) → `--env`. Valores `op://` resolvem-se em `BuildEnvFileValues` / ramos de secrets quando `OnePassword` não é nil.
4. **`unset`**: considerar chave ausente se não está no mapa do ficheiro **nem** em `LoadSecretKeys` para esse path.

## Armadilhas conhecidas (1Password / `op`)

- Erro **`isn't an item`** no `op item get`: `isItemNotFound` em `client.go` tem de reconhecer **`isn't`** (contracção), não só `is not an item`.
- **`op item edit` stdin**: remover **`vault`** do JSON antes de editar (`marshalItemJSONForOPEdit` em `itemjson.go`) para evitar inconsistência Private vs Employee.
- **`sudo mb`**: root usa **PATH** diferente — `mb` instalado em `~/.local/bin` ou via mise pode não existir; usar caminho absoluto ou não usar sudo no `mb` quando desnecessário.

## Documentação no repositório

- Guia principal: `docs/docs/guide/environment-variables.md` (inclui secção **1Password / `--secret-op`**).
- Tabela de comandos: `docs/docs/technical-reference/reference.md`.
- Recomendações de segurança: `docs/docs/guide/security.md`.

## Verificação

```bash
go test ./internal/app/envs/... ./internal/cli/envs/... ./internal/deps/... ./internal/infra/opcli/... -count=1
```

Ao alterar comportamento visível, alinhar **docs** e mensagens do Cobra com o guia acima.
