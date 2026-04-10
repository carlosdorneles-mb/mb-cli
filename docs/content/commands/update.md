---
sidebar_position: 4
---

# `mb update`

Atualiza, em sequência, plugins, ferramentas, o binário do CLI e pacotes do sistema operacional.

## Uso básico

```bash
mb update              # todas as fases
mb update --only-cli   # só o binário
mb update --only-plugins --only-cli  # fases combinadas
```

## Fases de atualização

| Fase | O que faz | Flag |
|---|---|---|
| **Plugins** | Atualiza plugins remotos (Git) para a versão mais recente | `--only-plugins` |
| **Tools** | Equivalente a `mb tools --update-all` (se o pacote `tools` estiver no cache com `update-all` no manifest) | `--only-tools` |
| **CLI** | Atualiza o binário do MB CLI para a release estável | `--only-cli` |
| **Sistema** | Delega em `mb machine update` (plugin shell: `brew`/`mas` no macOS, `apt`/`flatpak`/`snap` no Linux) | `--only-system` |

Sem nenhuma flag `--only-*`, todas as fases habilitadas são executadas.

## Flags

| Flag | Descrição |
|---|---|
| `--only-plugins` | Atualiza só os plugins |
| `--only-cli` | Atualiza só o binário do CLI |
| `--only-tools` | Atualiza ferramentas (equivalente a `mb tools --update-all`) |
| `--only-system` | Atualiza pacotes do SO (requer plugin `machine/update` no cache) |
| `--check-only` | Só com `--only-cli`: verifica release sem baixar |
| `--json` | Só com `--only-cli --check-only`: imprime JSON com versão local/remota |

## Atualização apenas do CLI

```bash
mb update --only-cli                    # baixa e instala a release mais nova
mb update --only-cli --check-only       # só verifica (código de saída 2 = há atualização)
mb update --only-cli --check-only --json  # saída JSON: {"localVersion":"...", "remoteVersion":"...", "updateAvailable":true}
```

> **Nota:** `--check-only` só funciona com binários de release oficial (com versão embutida via ldflags). Se compilou localmente ou usou `go install`, o comando indica que use `install.sh`.

## Flags `--only-*` condicionais

- **`--only-tools`** só aparece na ajuda quando o pacote `tools` está no cache com a flag `update-all` no `manifest.yaml`.
- **`--only-system`** só aparece quando o plugin `machine/update` está no cache.

Para detalhes sobre atualização de plugins, veja [Plugins](./plugins.md).
