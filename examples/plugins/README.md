# Plugins de exemplo (estrutura por diretórios)

A hierarquia de pastas define categorias e subcategorias. Cada pasta pode ter um `manifest.yaml`.

## Estrutura

| Caminho              | Invocação              |
|----------------------|------------------------|
| infra/ci/deploy      | `mb infra ci deploy` |
| infra/ci/lint        | `mb infra ci lint`   |
| infra/k8s/apply     | `mb infra k8s apply` |
| tools/hello          | `mb tools hello`     |

O path relativo à pasta de plugins vira o comando: `infra/ci/deploy` → `mb infra ci deploy`.

## Manifest

Toda categoria (e subcategoria) deve ter um `manifest.yaml` com pelo menos **description**; **command** e **readme** são opcionais. A description aparece na lista de comandos (ex.: `mb` mostra a descrição em "infra"; `mb infra` mostra a descrição em "ci"). O `--help` mostra apenas o help (Fang); o README é exibido somente com a flag `--readme`.

- **command** (opcional): nome do comando na CLI; default = nome do diretório.
- **description** (opcional mas recomendado para categorias): descrição (Short do Cobra).
- **readme** (opcional): arquivo (ex. README.md); quando existir, o comando ou categoria ganha a flag `--readme`, que renderiza o arquivo com glow.
- Com **entrypoint** e **type**: comando executável; qualquer flag/arg é repassada ao script.
- Sem entrypoint mas com **flags**: cada flag tem um entrypoint; sem flag → help.

## Como usar

1. Copie a árvore de plugins para o diretório de plugins do MB (ou use `make run-sandbox`, que já copia os exemplos):

   ```bash
   mkdir -p ~/.config/mb/plugins
   cp -r examples/plugins/infra examples/plugins/tools ~/.config/mb/plugins/
   chmod +x ~/.config/mb/plugins/infra/ci/deploy/run.sh
   chmod +x ~/.config/mb/plugins/infra/ci/lint/run.sh
   chmod +x ~/.config/mb/plugins/infra/k8s/apply/run.sh
   chmod +x ~/.config/mb/plugins/tools/hello/run.sh
   ```

2. Sincronize e liste:

   ```bash
   mb self sync
   mb plugins list
   mb infra ci deploy
   mb infra k8s apply
   mb tools hello
   ```

## Como ver o README

Se um comando ou categoria tiver `readme` no manifest (ex.: README.md), use a flag `--readme` para renderizar o arquivo com glow:

- **Categorias:** `mb tools --readme`, `mb infra ci --readme`, etc.
- **Comandos:** `mb infra ci deploy --readme`, `mb tools hello --readme` (quando o plugin tem readme configurado).
