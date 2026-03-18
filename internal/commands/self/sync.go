package self

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/deps"
	"mb/internal/helpers/shell"
	"mb/internal/system"
)

// RunSync rescans the plugins dir and registered local paths, upserts plugins and categories, and updates the plugin_sources registry.
// Used by both "mb self sync" and after plugins add/remove/update.
// log: gum log on stderr (warnings + optional success). If nil, warnings are dropped and success is not emitted.
func RunSync(ctx context.Context, deps deps.Dependencies, log *system.Logger, emitSuccess bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
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
	if log != nil {
		for _, w := range warnings {
			_ = log.Warn(ctx, "aviso: %s: %s", w.Path, w.Message)
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

	if emitSuccess && log != nil {
		_ = log.Info(ctx, "%d plugin(s) foram sincronizados", len(plugins))
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

func newSelfSyncCmd(deps deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "sync",
		Aliases: []string{"s"},
		Short:   "Rescaneia plugins e reconstrói o cache SQLite",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())
			return RunSync(ctx, deps, log, true)
		},
	}
}
