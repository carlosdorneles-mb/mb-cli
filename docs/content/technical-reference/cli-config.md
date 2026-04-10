---
sidebar_position: 4
---

# Configuração do CLI

O ficheiro **`~/.config/mb/config.yaml`** guarda opções **do próprio MB CLI**. Na **primeira execução**, se o ficheiro não existir, o MB **cria-o** com comentários e linhas em branco, **sem** preencher `docs_url` nem `update_repo` — os valores em falta vêm dos defaults do código na mesma.

Isto é **independente** dos ficheiros **`.env.*`** e de **`env.defaults`**, que servem para **variáveis de ambiente** mescladas em **`mb run`** e ao executar plugins.

## O que o `config.yaml` não faz

Só existem duas chaves lidas deste ficheiro (`docs_url` e `update_repo`). **Qualquer outra chave no YAML é ignorada** — por exemplo, `TESTCONFIG: dudu` **não** vira variável de ambiente e **`mb run`** não a verá.

Para variáveis como `TESTCONFIG`, use uma destas opções (ver [Variáveis de ambiente](../user-guide/environment-variables.md)):

- **`~/.config/mb/env.defaults`** — por exemplo `mb envs set TESTCONFIG dudu`
- **`./.env`** no diretório de trabalho atual
- **`mb run --env TESTCONFIG=dudu`** (ou `--env-file`)

## Chaves suportadas

| Chave | Obrigatória | Descrição |
|-------|-------------|-----------|
| `docs_url` | Não | URL base da documentação aberta com **`mb --doc`** (deve ser `http://` ou `https://` com host). Se omitir do ficheiro, usa-se o site público do projeto. |
| `update_repo` | Não | Repositório GitHub no formato **`owner/repo`** usado por **`mb update --only-cli`** (API de releases). Se omitir, usa-se o valor definido no build (`-X`) ou o repositório predefinido. |

## Comportamento

- **`docs_url` em falta** no YAML → URL de documentação predefinida (mesma do site oficial).
- **`update_repo` em falta** → `version.UpdateRepo` (ldflags) ou repositório upstream predefinido.
- **YAML inválido** ou valor inválido em `docs_url` / `update_repo` → o MB **não arranca** até corrigir o ficheiro.

## Exemplo

```yaml
# ~/.config/mb/config.yaml
docs_url: https://carlosdorneles-mb.github.io/mb-cli/
# Opcional: fork para update do CLI
# update_repo: minha-org/meu-fork-mb
```

Pode editar o ficheiro para definir `docs_url` e/ou `update_repo` quando quiser sobrescrever os defaults.
