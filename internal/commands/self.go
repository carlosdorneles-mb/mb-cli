package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/system"
	"mb/internal/ui"
)

func NewSelfCmd(deps Dependencies) *cobra.Command {
	selfCmd := &cobra.Command{
		Use:     "self",
		Short:   "Gerencia operações internas do MB CLI",
		GroupID: "commands",
	}

	selfCmd.AddCommand(newSelfSyncCmd(deps))
	selfCmd.AddCommand(newSelfEnvCmd(deps))
	return selfCmd
}

// RunSync rescans the plugins dir, upserts plugins and categories, and updates the plugin_sources registry.
// Used by both "mb self sync" and after plugins add/remove/update.
func RunSync(deps Dependencies, outSuccess func(string)) error {
	plugins, categories, err := deps.Scanner.Scan()
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		if err := deps.Store.UpsertPlugin(plugin); err != nil {
			return err
		}
	}

	if err := deps.Store.DeleteAllCategories(); err != nil {
		return err
	}
	for _, cat := range categories {
		if err := deps.Store.UpsertCategory(cat); err != nil {
			return err
		}
	}

	if err := updatePluginSourcesRegistry(deps, plugins, categories); err != nil {
		return err
	}

	if outSuccess != nil {
		outSuccess(fmt.Sprintf("synced %d plugin(s)", len(plugins)))
	}
	return nil
}

// updatePluginSourcesRegistry ensures plugin_sources has a row for each top-level dir under PluginsDir.
// Existing rows keep their git_url/ref_type/ref/version; new dirs get an empty row (manual install).
func updatePluginSourcesRegistry(deps Dependencies, plugins []cache.Plugin, categories []cache.Category) error {
	topLevelDirs := make(map[string]struct{})
	for _, p := range plugins {
		dir := firstPathSegment(p.CommandPath)
		if dir != "" {
			topLevelDirs[dir] = struct{}{}
		}
	}
	for _, c := range categories {
		dir := firstPathSegment(c.Path)
		if dir != "" {
			topLevelDirs[dir] = struct{}{}
		}
	}
	for dir := range topLevelDirs {
		existing, err := deps.Store.GetPluginSource(dir)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		if err := deps.Store.UpsertPluginSource(cache.PluginSource{InstallDir: dir}); err != nil {
			return err
		}
	}
	return nil
}

func firstPathSegment(path string) string {
	if path == "" {
		return ""
	}
	idx := strings.Index(path, "/")
	if idx == -1 {
		return path
	}
	return path[:idx]
}

func newSelfSyncCmd(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Rescaneia plugins e reconstrói o cache SQLite",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return RunSync(deps, func(msg string) {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(msg))
			})
		},
	}
}

func newSelfEnvCmd(deps Dependencies) *cobra.Command {
	selfEnvCmd := &cobra.Command{
		Use:   "env",
		Short: "Gerencia variáveis de ambiente padrão",
	}

	selfEnvCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Lista variáveis padrão",
		RunE: func(cmd *cobra.Command, _ []string) error {
			values, err := loadDefaultEnvValues(deps.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			keys := make([]string, 0, len(values))
			for key := range values {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", key, values[key])
			}
			return nil
		},
	})

	selfEnvCmd.AddCommand(&cobra.Command{
		Use:   "set KEY [VALUE]",
		Short: "Define uma variável padrão",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			var value string
			if len(args) == 2 {
				value = args[1]
			} else {
				input, err := system.Input(context.Background(), fmt.Sprintf("%s=", key))
				if err != nil {
					return err
				}
				value = input
			}

			values, err := loadDefaultEnvValues(deps.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			values[key] = value
			if err := saveDefaultEnvValues(deps.Runtime.DefaultEnvPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("saved %s", key)))
			return nil
		},
	})

	selfEnvCmd.AddCommand(&cobra.Command{
		Use:   "unset KEY",
		Short: "Remove uma variável padrão",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := loadDefaultEnvValues(deps.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			delete(values, args[0])
			if err := saveDefaultEnvValues(deps.Runtime.DefaultEnvPath, values); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("removed %s", args[0])))
			return nil
		},
	})

	return selfEnvCmd
}
