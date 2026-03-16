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

	if err := updatePluginSourcesRegistry(deps, plugins, categories); err != nil {
		return err
	}

	if outSuccess != nil {
		outSuccess(fmt.Sprintf("%d plugin(s) foram sincronizados", len(plugins)))
	}
	return nil
}

// updatePluginSourcesRegistry ensures plugin_sources has a row for each top-level dir under PluginsDir.
// Existing rows keep their git_url/ref_type/ref/version; new dirs get an empty row (manual install).
func updatePluginSourcesRegistry(deps config.Dependencies, plugins []cache.Plugin, categories []cache.Category) error {
	topLevelDirs := make(map[string]struct{})
	for _, p := range plugins {
		dir := FirstPathSegment(p.CommandPath)
		if dir != "" {
			topLevelDirs[dir] = struct{}{}
		}
	}
	for _, c := range categories {
		dir := FirstPathSegment(c.Path)
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
