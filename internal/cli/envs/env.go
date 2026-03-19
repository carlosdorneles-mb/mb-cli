package envs

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
)

// NewCmd returns the root "mb envs" command with list, set and unset subcommands.
func NewCmd(d deps.Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:     "envs",
		Aliases: []string{"e", "env"},
		Short:   "Gerencia variáveis de ambiente globais",
	}
	root.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	root.AddCommand(newListCmd(d))
	root.AddCommand(newSetCmd(d))
	root.AddCommand(newUnsetCmd(d))
	return root
}
