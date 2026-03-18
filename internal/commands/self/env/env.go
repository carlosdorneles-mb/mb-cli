package env

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
)

// NewCmd returns `mb self env` with list, set and unset subcommands.
func NewCmd(d deps.Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:     "env",
		Aliases: []string{"e", "envs", "settings"},
		Short:   "Gerencia variáveis de ambiente padrão",
	}
	root.AddGroup(&cobra.Group{ID: "commands", Title: "COMMANDOS"})

	root.AddCommand(newListCmd(d))
	root.AddCommand(newSetCmd(d))
	root.AddCommand(newUnsetCmd(d))
	return root
}
