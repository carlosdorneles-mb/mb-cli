package self

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/deps"
	"mb/internal/helpers/shell"
	plugpkg "mb/internal/plugins"
	"mb/internal/system"
)

// RunSync rescans the plugins dir and registered local paths, upserts plugins and categories, and updates the plugin_sources registry.
// Used by both "mb self sync" and after plugins add/remove/update.
// log: gum log on stderr (warnings + optional success). If nil, warnings are dropped and success is not emitted.
func RunSync(ctx context.Context, deps deps.Dependencies, log *system.Logger, emitSuccess bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Plugin-help debug (group_id / groups.yaml) sempre via system.Logger → gum log quando gum está no PATH.
	pluginHelpLog := log
	if pluginHelpLog == nil {
		quiet, verbose := false, false
		if deps.Runtime != nil {
			quiet, verbose = deps.Runtime.Quiet, deps.Runtime.Verbose
		}
		pluginHelpLog = system.NewLogger(quiet, verbose, os.Stderr)
	}
	deps.Scanner.DebugLog = func(msg string) { _ = pluginHelpLog.Debug(ctx, "%s", msg) }
	defer func() { deps.Scanner.DebugLog = nil }()

	plugins, categories, warnings, hgBatches, err := deps.Scanner.Scan()
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
		p, c, w, hg, err := deps.Scanner.ScanDir(src.LocalPath, src.InstallDir)
		if err != nil {
			return err
		}
		plugins = append(plugins, p...)
		categories = append(categories, c...)
		warnings = append(warnings, w...)
		hgBatches = append(hgBatches, hg...)
	}
	mergedHelp := plugpkg.MergeHelpGroupsGlobal(hgBatches, func(msg string) {
		_ = pluginHelpLog.Debug(ctx, "%s", msg)
	})
	if log != nil {
		for _, w := range warnings {
			_ = log.Warn(ctx, "aviso: %s: %s", w.Path, w.Message)
		}
	}

	if err := checkPluginPathCollisions(plugins); err != nil {
		return err
	}

	validGroupIDs := make(map[string]struct{}, len(mergedHelp))
	for _, g := range mergedHelp {
		validGroupIDs[g.ID] = struct{}{}
	}
	normalizePluginGroupIDs(plugins, validGroupIDs, func(msg string) {
		_ = pluginHelpLog.Debug(ctx, "%s", msg)
	})
	normalizeCategoryGroupIDs(categories, validGroupIDs, func(msg string) {
		_ = pluginHelpLog.Debug(ctx, "%s", msg)
	})

	if err := deps.Store.DeleteAllPlugins(); err != nil {
		return err
	}
	if err := deps.Store.DeleteAllPluginHelpGroups(); err != nil {
		return err
	}
	for _, g := range mergedHelp {
		if err := deps.Store.UpsertPluginHelpGroup(cache.PluginHelpGroup{GroupID: g.ID, Title: g.Title}); err != nil {
			return err
		}
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

func normalizeCategoryGroupIDs(categories []cache.Category, valid map[string]struct{}, debug func(string)) {
	for i := range categories {
		c := &categories[i]
		if !strings.Contains(c.Path, "/") {
			c.GroupID = ""
			continue
		}
		if c.GroupID == "" {
			continue
		}
		if _, ok := valid[c.GroupID]; !ok {
			if debug != nil {
				debug(fmt.Sprintf("plugin help: category_path=%q group_id=%q não cadastrado em nenhum groups.yaml; usando COMANDOS", c.Path, c.GroupID))
			}
			c.GroupID = ""
		}
	}
}

func normalizePluginGroupIDs(plugins []cache.Plugin, valid map[string]struct{}, debug func(string)) {
	for i := range plugins {
		p := &plugins[i]
		if !strings.Contains(p.CommandPath, "/") {
			p.GroupID = ""
			continue
		}
		if p.GroupID == "" {
			continue
		}
		if _, ok := valid[p.GroupID]; !ok {
			if debug != nil {
				debug(fmt.Sprintf("plugin help: command_path=%q group_id=%q não cadastrado em nenhum groups.yaml; usando COMANDOS", p.CommandPath, p.GroupID))
			}
			p.GroupID = ""
		}
	}
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
