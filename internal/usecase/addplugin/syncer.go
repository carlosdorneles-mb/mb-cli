package addplugin

import (
	"context"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
	"mb/internal/shared/syncdiff"
)

// Syncer wraps the RunSync logic so it can be injected as a dependency
// rather than called as a package-level function.
type Syncer struct{}

// NewSyncer creates a Syncer.
func NewSyncer() *Syncer {
	return &Syncer{}
}

// SyncerOptions configures RunSync behaviour.
type SyncerOptions struct {
	EmitSuccess bool
	PostSync    func(context.Context) error
}

// Run rescans PluginsDir and each plugin_sources.local_path tree, merges
// results, and replaces plugin, category, and help-group rows in SQLite.
func (*Syncer) Run(
	ctx context.Context,
	rt Runtime,
	store ports.PluginSyncStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	log Logger,
	opts SyncerOptions,
) (SyncReport, error) {
	runCtx := ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	beforePlugins, err := store.ListPlugins()
	if err != nil {
		return SyncReport{}, err
	}
	beforeByKey := syncdiff.PluginsByCommandKey(beforePlugins)

	var debugFn syncdiff.DebugFunc
	if log != nil {
		debugFn = func(msg string) { _ = log.Debug(runCtx, "%s", msg) }
	} else {
		debugFn = func(string) {}
	}

	scanner.SetDebugLog(debugFn)
	defer func() { scanner.SetDebugLog(nil) }()

	pluginsList, categories, warnings, hgBatches, err := scanner.Scan()
	if err != nil {
		return SyncReport{}, err
	}
	if _, err := shell.EnsureShellHelpers(rt.ConfigDir); err != nil {
		return SyncReport{}, err
	}
	sources, err := store.ListPluginSources()
	if err != nil {
		return SyncReport{}, err
	}
	for _, src := range sources {
		if src.LocalPath == "" {
			continue
		}
		p, c, w, hg, err := scanner.ScanDir(src.LocalPath, src.InstallDir)
		if err != nil {
			return SyncReport{}, err
		}
		pluginsList = append(pluginsList, p...)
		categories = append(categories, c...)
		warnings = append(warnings, w...)
		hgBatches = append(hgBatches, hg...)
	}
	mergedHelp := plugin.MergeHelpGroupsGlobal(hgBatches, debugFn)
	if log != nil {
		for _, w := range warnings {
			_ = log.Warn(runCtx, "aviso: %s: %s", w.Path, w.Message)
		}
	}

	afterByKey := syncdiff.PluginsByCommandKey(pluginsList)
	removedKeys := syncdiff.DiffRemovedKeys(beforeByKey, afterByKey)

	if err := syncdiff.CheckPluginPathCollisions(pluginsList); err != nil {
		return SyncReport{}, err
	}

	sd := syncdiff.EmitDiff(runCtx, &loggerAdapter{log}, beforeByKey, pluginsList, removedKeys)

	validGroupIDs := make(map[string]struct{}, len(mergedHelp))
	for _, g := range mergedHelp {
		validGroupIDs[g.ID] = struct{}{}
	}
	syncdiff.NormalizePluginGroupIDs(pluginsList, validGroupIDs, debugFn)
	syncdiff.NormalizeCategoryGroupIDs(categories, validGroupIDs, debugFn)

	if err := store.DeleteAllPlugins(); err != nil {
		return SyncReport{}, err
	}
	if err := store.DeleteAllPluginHelpGroups(); err != nil {
		return SyncReport{}, err
	}
	for _, g := range mergedHelp {
		if err := store.UpsertPluginHelpGroup(
			plugin.PluginHelpGroup{GroupID: g.ID, Title: g.Title},
		); err != nil {
			return SyncReport{}, err
		}
	}
	for _, plg := range pluginsList {
		if err := store.UpsertPlugin(plg); err != nil {
			return SyncReport{}, err
		}
	}

	if err := store.DeleteAllCategories(); err != nil {
		return SyncReport{}, err
	}
	for _, cat := range categories {
		if err := store.UpsertCategory(cat); err != nil {
			return SyncReport{}, err
		}
	}

	report := SyncReport{
		Added:     sd.Added,
		Updated:   sd.Updated,
		Removed:   sd.Removed,
		AnyChange: sd.AnyChange,
	}

	if opts.EmitSuccess && log != nil && !report.AnyChange {
		_ = log.Info(runCtx, "Nenhum comando alterado; cache atualizado.")
	}

	if opts.PostSync != nil {
		if err := opts.PostSync(runCtx); err != nil && log != nil {
			_ = log.Warn(runCtx, "autocompletar: %v", err)
		}
	}
	return report, nil
}

// loggerAdapter adapts addplugin.Logger to syncdiff.Logger.
type loggerAdapter struct{ log Logger }

func (a *loggerAdapter) Info(ctx context.Context, msg string, args ...any) error {
	if a.log == nil {
		return nil
	}
	return a.log.Info(ctx, msg, args...)
}
func (a *loggerAdapter) Warn(ctx context.Context, msg string, args ...any) error {
	if a.log == nil {
		return nil
	}
	return a.log.Warn(ctx, msg, args...)
}
func (a *loggerAdapter) Debug(ctx context.Context, msg string, args ...any) error {
	if a.log == nil {
		return nil
	}
	return a.log.Debug(ctx, msg, args...)
}
