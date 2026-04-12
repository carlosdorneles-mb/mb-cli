package envs

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/usecase/envs"
)

// NewCmd returns the root "mb envs" command with list, set and unset subcommands.
func NewCmd(listSvc *envs.ListService, d deps.Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:     "envs",
		Aliases: []string{"e", "env"},
		Short:   "Gerencia variáveis de ambiente globais",
	}
	root.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	root.AddCommand(newListCmd(listSvc, d.Runtime.ConfigDir))
	root.AddCommand(newVaultsCmd(d))
	root.AddCommand(newSetCmd(d))
	root.AddCommand(newUnsetCmd(d))
	return root
}
