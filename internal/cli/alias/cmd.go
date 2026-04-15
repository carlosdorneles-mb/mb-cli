package alias

import (
	"os"

	"github.com/spf13/cobra"

	"mb/internal/deps"
)

// NewCmd returns the `mb alias` command tree.
func NewCmd(d deps.Dependencies) *cobra.Command {
	var shellFlag string

	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Aliases de comando no shell e mb run",
		Long: `Define atalhos de comando persistentes no MB CLI e gera scripts para o seu shell.

Ao usar um subcomando, o MB garante no perfil do shell (caminho padrão para o shell detectado
ou indicado com --shell) um bloco idempotente (marcadores # mb-cli user aliases BEGIN/END) que
carrega esses scripts. Para remover só o bloco do perfil, edite o arquivo manualmente.

Chamar um alias no terminal não aplica o ambiente mesclado do MB; use mb run <nome> para isso
(e opcionalmente --vault no mb alias set para associar um vault ao slot usado em mb run).

Aliases em mbcli.yaml (mb alias set/unset --mbcli-yaml) aparecem no mb alias list e são resolvidos
no mb run; não viram atalhos no shell — só o mb run aplica o ambiente mesclado.`,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			if skipAliasProfileForHelp(os.Args[1:]) {
				return nil
			}
			// Bare `mb alias` only prints help — do not touch the profile first.
			if c.Name() == "alias" && len(args) == 0 {
				return nil
			}
			return ensureShellProfileForAliases(d, c, EnsureProfileOptions{
				ShellFlag: shellFlag,
			})
		},
	}

	cmd.PersistentFlags().StringVar(
		&shellFlag, "shell", "",
		"Shell alvo (bash, zsh, fish, powershell); por padrão detecta via SHELL",
	)

	cmd.AddCommand(newSetCmd(d))
	cmd.AddCommand(newListCmd(d))
	cmd.AddCommand(newUnsetCmd(d))
	return cmd
}
