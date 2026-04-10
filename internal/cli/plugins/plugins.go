package plugins

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/usecase/addplugin"
)

func NewPluginsCmd(svc *addplugin.Service, deps deps.Dependencies) *cobra.Command {
	pluginsCmd := &cobra.Command{
		Use:     "plugins",
		Aliases: []string{"plugin", "p", "extensions", "e"},
		Short:   "Gerencia plugins no CLI, adicionar, listar, remover e atualizar",
	}
	pluginsCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	addCmd := newPluginsAddCmd(svc, deps)
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
	syncCmd := newPluginsSyncCmd(deps)
	syncCmd.GroupID = "commands"
	pluginsCmd.AddCommand(syncCmd)
	return pluginsCmd
}
