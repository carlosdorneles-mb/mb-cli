# Helpers de shell

Scripts de shell embutidos no MB CLI e copiados para `~/.config/mb/lib/shell` no `mb self sync`. Os plugins recebem a variável **`MB_HELPERS_PATH`** apontando para esse diretório e podem carregá-los (ex.: `. "$MB_HELPERS_PATH/all.sh"`) para usar funções como `log` que respeitam as flags do CLI.

**Documentação:** [Helpers de shell](../../../docs/docs/helpers-shell.md) — como carregar nos plugins e lista de helpers disponíveis.

Helpers atuais:

- `all.sh`
- `log.sh`
- `memory.sh`
- `string.sh`
- `kubernetes.sh`

---

## Adicionar um novo arquivo `.sh`

1. **Criar o arquivo** em `internal/helpers/shell/` (ex.: `meu.sh`). O embed usa `*.sh`, então o novo arquivo passa a ser incluído automaticamente.
2. **Incluir no `all.sh`** — adicionar a linha que faz o source do novo helper, por exemplo:  
   `. "${MB_HELPERS_PATH}/meu.sh"`
3. **Atualizar a documentação** em `docs/docs/helpers-shell.md`: descrever o novo helper na seção "Helpers disponíveis" e, se for carregável isoladamente, na seção "Como carregar".
