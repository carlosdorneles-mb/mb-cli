package plugincmd

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/commands/self"
	mbplugins "mb/internal/plugins"
	"mb/internal/system"
)

func newPluginsListCmd(deps config.Dependencies) *cobra.Command {
	var checkUpdates bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista plugins instalados (name, command, description, version, url)",
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginList, err := deps.Store.ListPlugins()
			if err != nil {
				return err
			}
			sources, err := deps.Store.ListPluginSources()
			if err != nil {
				return err
			}
			sourceByDir := make(map[string]*cache.PluginSource)
			for i := range sources {
				sourceByDir[sources[i].InstallDir] = &sources[i]
			}

			sort.Slice(pluginList, func(i, j int) bool {
				return pluginList[i].CommandPath < pluginList[j].CommandPath
			})

			rows := make([][]string, 0, len(pluginList))
			for _, p := range pluginList {
				installDir := self.FirstPathSegment(p.CommandPath)
				src := sourceByDir[installDir]
				name := installDir
				version := "-"
				origem := "-"
				url := "-"
				if src != nil {
					version = src.Version
					if src.LocalPath != "" {
						origem = "local"
						url = src.LocalPath
					} else if src.GitURL != "" {
						origem = "remoto"
						url = src.GitURL
					}
				}

				updateAvail := ""
				if checkUpdates && src != nil && src.GitURL != "" && src.LocalPath == "" {
					dir := filepath.Join(deps.Runtime.PluginsDir, installDir)
					if mbplugins.IsGitRepo(dir) {
						if src.RefType == "tag" {
							_ = mbplugins.FetchTags(cmd.Context(), dir)
							tags, _ := mbplugins.ListLocalTags(dir)
							for _, t := range tags {
								if _, newer := mbplugins.NewerTag(src.Ref, t); newer {
									updateAvail = "sim"
									break
								}
							}
						} else {
							updateAvail = "-"
						}
					}
				}

				rows = append(rows, []string{name, p.CommandPath, p.Description, version, origem, url, updateAvail})
			}

			headers := []string{"NOME", "COMANDO", "DESCRIÇÃO", "VERSÃO", "ORIGEM", "URL", "ATUALIZAR"}
			if !checkUpdates {
				headers = []string{"NOME", "COMANDO", "DESCRIÇÃO", "VERSÃO", "ORIGEM", "URL"}
				for i := range rows {
					rows[i] = rows[i][:6]
				}
			}
			return system.Table(context.Background(), headers, rows, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&checkUpdates, "check-updates", false, "Verifica se há atualização disponível para cada plugin")
	return cmd
}
