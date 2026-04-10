# Helpers de shell

Scripts de shell embutidos no MB CLI e **sempre** materializados em `~/.config/mb/lib/shell` em cada `mb plugins sync`: o conteúdo em disco é substituído pelo que está embebido no binário, e ficheiros `*.sh` órfãos (já não presentes no embed) são removidos. Os plugins recebem a variável **`MB_HELPERS_PATH`** apontando para esse diretório e podem carregá-los (ex.: `. "$MB_HELPERS_PATH/all.sh"`) para usar funções como `log` que respeitam as flags do CLI.

**Documentação:** [Helpers de shell](../../../../docs/docs/technical-reference/helpers-shell.md) — como carregar nos plugins e lista de helpers disponíveis. Variáveis de contexto `MB_CTX_*` (runtime do CLI): [Contexto de invocação de plugins](../../../../docs/docs/technical-reference/plugin-invocation-context.md).

Helpers adicionais embebidos incluem **`mbcli-yaml.sh`** (leitura/escrita de `mbcli.yaml` no projeto com `yq`).

---

## Adicionar um novo arquivo `.sh`

1. **Criar o arquivo** em `internal/infra/shellhelpers/` (ex.: `meu.sh`). O embed usa `*.sh`, então o novo arquivo passa a ser incluído automaticamente.
2. **Incluir no `all.sh`** — adicionar a linha que faz o source do novo helper, ex.: `. "${MB_HELPERS_PATH}/meu.sh"`.
3. **Atualizar a documentação** em `docs/docs/technical-reference/helpers-shell.md`: descrever o novo helper na seção "Helpers disponíveis" e, se for carregável isoladamente, na seção "Como carregar".

---

## Atualizar o helpers no CLI

Recompile o `mb` (ou use `make run` / `make run-local`) e execute `mb plugins sync` — não há atalho por checksum: cada sync alinha `lib/shell` com o embed da versão do binário em execução.
