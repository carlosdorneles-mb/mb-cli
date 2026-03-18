package plugins

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
)

func NewPluginsCmd(deps deps.Dependencies) *cobra.Command {
	pluginsCmd := &cobra.Command{
		Use:     "plugins",
		Short:   "Gerencia plugins instalados (add, list, remove, update)",
		GroupID: "commands",
	}
	pluginsCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	addCmd := newPluginsAddCmd(deps)
	addCmd.GroupID = "commands"
	pluginsCmd.AddCommand(addCmd)
	listCmd := newPluginsListCmd(deps)
	listCmd.GroupID = "commands"
	pluginsCmd.AddCommand(listCmd)
	removeCmd := newPluginsRemoveCmd(deps)
	removeCmd.GroupID = "commands"
	pluginsCmd.AddCommand(removeCmd)
	updateCmd := newPluginsUpdateCmd(deps)
	updateCmd.GroupID = "commands"
	pluginsCmd.AddCommand(updateCmd)
	return pluginsCmd
}
