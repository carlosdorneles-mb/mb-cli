package plugins

import (
	"context"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
	"mb/internal/shared/syncdiff"
)

// SyncOptions configures RunSync behaviour.
type SyncOptions struct {
	// EmitSuccess logs a short summary when true (e.g. plugins sync command).
	EmitSuccess bool
	// PostSync runs after a successful SQLite refresh.
	PostSync func(context.Context) error
}

// SyncReport summarizes plugin command changes detected during sync (leaf commands only).
type SyncReport = syncdiff.SyncReport

// RunSync rescans PluginsDir and each plugin_sources.local_path tree, merges results, and replaces plugin, category, and help-group rows in SQLite. It does not mutate plugin_sources (add/remove/update handle that).
// Used by mb plugins sync and after plugins add/remove/update.
// log: warnings and success messages. If nil, warnings are dropped and success is not emitted.
func RunSync(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginSyncStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	log ports.Logger,
	opts SyncOptions,
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
	scanner.SetDebugLog(func(msg string) { _ = logDebugOrNil(log, runCtx, msg) })
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
	mergedHelp := plugin.MergeHelpGroupsGlobal(hgBatches, func(msg string) {
		_ = logDebugOrNil(log, runCtx, msg)
	})
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

	report := syncdiff.EmitDiff(runCtx, log, beforeByKey, pluginsList, removedKeys)

	validGroupIDs := make(map[string]struct{}, len(mergedHelp))
	for _, g := range mergedHelp {
		validGroupIDs[g.ID] = struct{}{}
	}
	syncdiff.NormalizePluginGroupIDs(pluginsList, validGroupIDs, func(msg string) {
		_ = logDebugOrNil(log, runCtx, msg)
	})
	syncdiff.NormalizeCategoryGroupIDs(categories, validGroupIDs, func(msg string) {
		_ = logDebugOrNil(log, runCtx, msg)
	})

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

// logDebugOrNil safely calls log.Debug or does nothing if log is nil.
func logDebugOrNil(log ports.Logger, ctx context.Context, msg string) error {
	if log == nil {
		return nil
	}
	return log.Debug(ctx, "%s", msg)
}
