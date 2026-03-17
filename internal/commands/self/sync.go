package self

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/helpers/shell"
	"mb/internal/ui"
)

// RunSync rescans the plugins dir and registered local paths, upserts plugins and categories, and updates the plugin_sources registry.
// Used by both "mb self sync" and after plugins add/remove/update.
// outWarnings: if non-nil, validation warnings (skipped plugins) are written here.
func RunSync(deps config.Dependencies, outSuccess func(string), outWarnings io.Writer) error {
	plugins, categories, warnings, err := deps.Scanner.Scan()
	if err != nil {
		return err
	}
	if _, err := shell.EnsureShellHelpers(deps.Runtime.ConfigDir); err != nil {
		return err
	}
	sources, err := deps.Store.ListPluginSources()
	if err != nil {
		return err
	}
	for _, src := range sources {
		if src.LocalPath == "" {
			continue
		}
		p, c, w, err := deps.Scanner.ScanDir(src.LocalPath, src.InstallDir)
		if err != nil {
			return err
		}
		plugins = append(plugins, p...)
		categories = append(categories, c...)
		warnings = append(warnings, w...)
	}
	if outWarnings != nil {
		for _, w := range warnings {
			fmt.Fprintf(outWarnings, "aviso: %s: %s\n", w.Path, w.Message)
		}
	}

	if err := checkPluginPathCollisions(plugins); err != nil {
		return err
	}

	if err := deps.Store.DeleteAllPlugins(); err != nil {
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

	if outSuccess != nil {
		outSuccess(fmt.Sprintf("%d plugin(s) foram sincronizados", len(plugins)))
	}
	return nil
}

func checkPluginPathCollisions(plugins []cache.Plugin) error {
	seen := make(map[string]string)
	for _, p := range plugins {
		key := p.CommandPath
		if key == "" {
			key = p.CommandName
		}
		if prevDir, ok := seen[key]; ok {
			if prevDir != p.PluginDir {
				return fmt.Errorf("conflito de plugins: o caminho de comando %q está definido em dois pacotes (%s e %s). Remova ou ajuste uma das fontes", key, prevDir, p.PluginDir)
			}
			continue
		}
		seen[key] = p.PluginDir
	}
	return nil
}

func newSelfSyncCmd(deps config.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Rescaneia plugins e reconstrói o cache SQLite",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return RunSync(deps, func(msg string) {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(msg))
			}, cmd.ErrOrStderr())
		},
	}
}
