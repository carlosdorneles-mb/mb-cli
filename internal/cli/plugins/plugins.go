package plugins

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/usecase/addplugin"
	"mb/internal/usecase/plugins"
)

func NewPluginsCmd(
	addSvc *addplugin.Service,
	syncSvc *plugins.SyncService,
	rmSvc *plugins.RemoveService,
	upSvc *plugins.UpdateService,
	deps deps.Dependencies,
) *cobra.Command {
	pluginsCmd := &cobra.Command{
		Use:     "plugins",
		Aliases: []string{"plugin", "p", "extensions", "e"},
		Short:   "Gerencia plugins no CLI, adicionar, listar, remover e atualizar",
	}
	pluginsCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	addCmd := newPluginsAddCmd(addSvc, deps)
	addCmd.GroupID = "commands"
	pluginsCmd.AddCommand(addCmd)
	listCmd := newPluginsListCmd(deps)
	listCmd.GroupID = "commands"
	pluginsCmd.AddCommand(listCmd)
	removeCmd := newPluginsRemoveCmd(rmSvc, deps)
	removeCmd.GroupID = "commands"
	pluginsCmd.AddCommand(removeCmd)
	updateCmd := newPluginsUpdateCmd(upSvc, deps)
	updateCmd.GroupID = "commands"
	pluginsCmd.AddCommand(updateCmd)
	syncCmd := newPluginsSyncCmd(syncSvc, deps)
	syncCmd.GroupID = "commands"
	pluginsCmd.AddCommand(syncCmd)
	return pluginsCmd
}
