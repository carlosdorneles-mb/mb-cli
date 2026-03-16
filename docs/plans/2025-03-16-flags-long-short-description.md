# Plano: Long/short e descrição por flag

**Objetivo:** (1) Permitir que uma mesma flag tenha forma longa e curta (ex.: `--deploy` e `-d`). (2) Permitir descrição por flag para aparecer no help do comando (`mb tools do --help`).

---

## Alterações em relação ao plano anterior

- Incluir **descrição por flag**: campo opcional no `FlagDef` (ex.: `description`) usado como *usage* ao registrar a flag no Cobra, para aparecer no `--help`.

---

## 1. Schema do manifesto (FlagDef)

**Arquivo:** [internal/plugins/manifest.go](internal/plugins/manifest.go)

- Adicionar em `FlagDef`:
  - `Short string \`yaml:"short"\`` — opcional; uma letra para a forma `-x` (usada junto com o nome longo = chave do map).
  - `Description string \`yaml:"description"\`` — opcional; texto exibido no help do comando para essa flag (Cobra *usage*).

Exemplo YAML:

```yaml
flags:
  deploy:
    type: long
    short: d
    description: "Faz o deploy do ambiente"
    entrypoint: deploy.sh
  rollback:
    type: long
    short: r
    description: "Reverte o último deploy"
    entrypoint: rollback.sh
```

---

## 2. Registro de flags (dynamic.go)

**Arquivo:** [internal/commands/dynamic.go](internal/commands/dynamic.go)

- Definir variável de uso: `usage := def.Description` (se vazio, usar `""`).
- No loop que registra as flags:
  - Se `def.Short != ""` e um único rune: `cmd.Flags().BoolP(name, def.Short, false, usage)`.
  - Senão, manter lógica atual por `def.Type`, passando `usage` no último argumento em vez de `""`:
    - `cmd.Flags().Bool(name, false, usage)`
    - `cmd.Flags().BoolP(name, name, false, usage)` (quando type short e nome 1 char).

Assim tanto `--deploy`/`-d` quanto a descrição aparecem no `mb tools do --help`.

---

## 3. Cache / FlagsJSON

O cache guarda o JSON do map `flags` (já serializa a struct `FlagDef`). Ao adicionar `Short` e `Description`, o JSON passará a incluí-los; não é necessário mudar o schema do cache, só garantir que o scanner não filtre esses campos (o `json.Marshal(manifest.Flags)` já inclui todos os campos exportados).

---

## 4. Documentação e exemplo

- **creating-plugins.md:** Na seção de flags, documentar:
  - `short` (opcional): uma letra para `-x`; a chave do map é o nome longo (`--nome`).
  - `description` (opcional): texto mostrado no help do comando para essa flag.
- **examples/plugins/tools/do/manifest.yaml:** Usar `short` e `description` em pelo menos uma flag (ex.: deploy com `short: d` e `description: "Faz o deploy"`).

---

## 5. Resumo de arquivos

| Arquivo | Ação |
|---------|------|
| [internal/plugins/manifest.go](internal/plugins/manifest.go) | Adicionar `Short` e `Description` em `FlagDef`. |
| [internal/commands/dynamic.go](internal/commands/dynamic.go) | Registrar flags com `BoolP` quando `def.Short` for 1 rune; passar `def.Description` como usage em todos os registros. |
| Validação (scanner ou dynamic) | Opcional: `short` com 1 rune; shorts únicos por comando. |
| [docs-site/docs/creating-plugins.md](docs-site/docs/creating-plugins.md) | Documentar `short` e `description` por flag. |
| [examples/plugins/tools/do/manifest.yaml](examples/plugins/tools/do/manifest.yaml) | Exemplo com `short` e `description`. |
| Testes | Cobrir flag com short + description; retrocompat (sem short/description). |

---

## 6. Ordem sugerida de implementação

1. Adicionar `Short` e `Description` em `FlagDef` (manifest.go).
2. Em dynamic.go: usar `def.Description` como usage e registrar com `BoolP(name, def.Short, ...)` quando `def.Short` for 1 rune; senão manter tipo long/short atual, passando usage.
3. Validação opcional de short (1 rune, único).
4. Testes (unit ou integração).
5. Doc (creating-plugins) e exemplo (tools/do manifest).
