package plugincmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/system"
	"mb/internal/shared/ui"
)

// ensureNestedPluginHelpGroups registers help groups on cmd so Cobra children with matching GroupID validate.
// Cobra requires every child GroupID to exist on the direct parent. Category commands
// get this before AddCommand; a manifest leaf (e.g. tools) later used as parent for tools/bruno
// must receive the same groups before AddCommand.
func ensureNestedPluginHelpGroups(cmd *cobra.Command, helpGroups []sqlite.PluginHelpGroup) {
	if cmd == nil {
		return
	}
	if !cmd.ContainsGroup("commands") {
		cmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})
	}
	for _, hg := range helpGroups {
		if hg.GroupID == "" || cmd.ContainsGroup(hg.GroupID) {
			continue
		}
		cmd.AddGroup(&cobra.Group{ID: hg.GroupID, Title: hg.Title})
	}
}

// Attach registers plugin commands from the cache under root.
func Attach(root *cobra.Command, d deps.Dependencies) {
	pluginList, err := d.Store.ListPlugins()
	if err != nil {
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderError(err.Error()))
		fmt.Fprintln(
			root.ErrOrStderr(),
			ui.RenderInfo("Execute `mb plugins sync` para reconstruir o cache de plugins."),
		)
		return
	}
	if len(pluginList) == 0 {
		fmt.Fprintln(
			root.ErrOrStderr(),
			ui.RenderInfo("Não há plugins em cache. Execute `mb plugins sync` primeiro."),
		)
		return
	}

	categoryList, _ := d.Store.ListCategories()
	categoriesByPath := make(map[string]sqlite.Category)
	for _, c := range categoryList {
		categoriesByPath[c.Path] = c
	}

	sources, _ := d.Store.ListPluginSources()
	sourceByDir := make(map[string]*sqlite.PluginSource)
	for i := range sources {
		sourceByDir[sources[i].InstallDir] = &sources[i]
	}

	sort.Slice(pluginList, func(i, j int) bool {
		return pluginList[i].CommandPath < pluginList[j].CommandPath
	})

	helpGroups, _ := d.Store.ListPluginHelpGroups()
	registeredHelp := make(map[string]struct{})
	for _, hg := range helpGroups {
		root.AddGroup(&cobra.Group{ID: hg.GroupID, Title: hg.Title})
		registeredHelp[hg.GroupID] = struct{}{}
	}
	// Inconsistências de group_id no cache: debug via gum log (mesmo pipeline que mb plugins sync -v).
	dbgLog := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, root.ErrOrStderr())
	globalShorts := persistentShorthandSet(root)

	type node struct {
		cmd     *cobra.Command
		plugins []sqlite.Plugin
	}
	byPath := map[string]*node{}

	for _, plugin := range pluginList {
		segments := strings.Split(plugin.CommandPath, "/")
		if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
			segments = []string{plugin.CommandName}
		}

		var parent = root
		for i := 0; i < len(segments)-1; i++ {
			seg := segments[i]
			pathSoFar := strings.Join(segments[:i+1], "/")
			if byPath[pathSoFar] == nil {
				cat := categoriesByPath[pathSoFar]
				short := seg + " COMANDOS"
				if cat.Description != "" {
					short = cat.Description
				}
				categoryCmd := &cobra.Command{
					Use:   seg,
					Short: short,
				}
				if parent == root {
					categoryCmd.GroupID = "plugin_commands"
				} else if cat.GroupID != "" {
					if _, ok := registeredHelp[cat.GroupID]; ok {
						categoryCmd.GroupID = cat.GroupID
					} else {
						_ = dbgLog.Debug(
							context.Background(),
							"plugin help: category_path=%q group_id=%q não registado no cache; usando COMANDOS",
							pathSoFar,
							cat.GroupID,
						)
						categoryCmd.GroupID = "commands"
					}
				} else {
					categoryCmd.GroupID = "commands"
				}
				if cat.ReadmePath != "" {
					categoryCmd.Flags().BoolP("readme", "r", false, readmeFlagDesc)
					categoryCmd.RunE = func(cmd *cobra.Command, _ []string) error {
						if cmd.Flags().Lookup("readme").Changed {
							return runReadmeWithGlow(cat.ReadmePath)
						}
						cmd.Help()
						return nil
					}
				}
				setHelpFang(categoryCmd)
				categoryCmd.Hidden = cat.Hidden
				if parent != root {
					ensureNestedPluginHelpGroups(parent, helpGroups)
				}
				ensureNestedPluginHelpGroups(categoryCmd, helpGroups)
				parent.AddCommand(categoryCmd)
				byPath[pathSoFar] = &node{cmd: categoryCmd}
			}
			parent = byPath[pathSoFar].cmd
		}

		pathSoFar := plugin.CommandPath
		if byPath[pathSoFar] != nil {
			continue
		}
		src := plugins.SourceForPlugin(plugin, sources, d.Runtime.PluginsDir)
		pluginRoot := plugin.PluginDir
		if pluginRoot == "" {
			installDir := plugins.FirstPathSegment(plugin.CommandPath)
			s := sourceByDir[installDir]
			pluginRoot = filepath.Join(d.Runtime.PluginsDir, installDir)
			if s != nil && s.LocalPath != "" {
				pluginRoot = s.LocalPath
			}
		}
		isLocal := src != nil && src.LocalPath != ""
		leafCmd := newLeafCommand(
			plugin.CommandName,
			plugin,
			d,
			pluginRoot,
			isLocal,
			dbgLog,
			globalShorts,
		)
		leafCmd.Hidden = plugin.Hidden
		if parent == root {
			leafCmd.GroupID = "plugin_commands"
		} else if plugin.GroupID != "" {
			if _, ok := registeredHelp[plugin.GroupID]; ok {
				leafCmd.GroupID = plugin.GroupID
			} else {
				_ = dbgLog.Debug(
					context.Background(),
					"plugin help: command_path=%q group_id=%q não registado no cache; usando COMANDOS",
					plugin.CommandPath,
					plugin.GroupID,
				)
				leafCmd.GroupID = "commands"
			}
		} else {
			leafCmd.GroupID = "commands"
		}
		if parent != root {
			ensureNestedPluginHelpGroups(parent, helpGroups)
		}
		parent.AddCommand(leafCmd)
		byPath[pathSoFar] = &node{cmd: leafCmd, plugins: []sqlite.Plugin{plugin}}
	}
}
