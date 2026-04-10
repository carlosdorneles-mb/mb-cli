package plugins

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
)

// SyncOptions configures RunSync behaviour.
type SyncOptions struct {
	// EmitSuccess logs a short summary when true (e.g. plugins sync command): one line when
	// nothing changed, or no extra bulk line when per-command diff lines were emitted.
	EmitSuccess bool
	// NoRemove keeps SQLite rows for commands that disappeared from scanned packages (orphans).
	NoRemove bool
	// PostSync runs after a successful SQLite refresh. Errors are logged but do not fail RunSync.
	PostSync func(context.Context) error
}

// SyncReport summarizes plugin command changes detected during sync (leaf commands only).
type SyncReport struct {
	Added     int
	Updated   int
	Removed   int
	AnyChange bool
}

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
	beforeByKey := pluginsByCommandKey(beforePlugins)
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

	afterByKey := pluginsByCommandKey(pluginsList)
	removedKeys := diffRemovedKeys(beforeByKey, afterByKey)

	if opts.NoRemove && len(removedKeys) > 0 {
		for _, k := range removedKeys {
			pluginsList = append(pluginsList, beforeByKey[k])
		}
	}

	if err := checkPluginPathCollisions(pluginsList); err != nil {
		return SyncReport{}, err
	}

	report := emitPluginSyncDiff(runCtx, log, beforeByKey, pluginsList, removedKeys, opts.NoRemove)

	validGroupIDs := make(map[string]struct{}, len(mergedHelp))
	for _, g := range mergedHelp {
		validGroupIDs[g.ID] = struct{}{}
	}
	normalizePluginGroupIDs(pluginsList, validGroupIDs, func(msg string) {
		_ = logDebugOrNil(log, runCtx, msg)
	})
	normalizeCategoryGroupIDs(categories, validGroupIDs, func(msg string) {
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

func pluginCommandKey(p plugin.Plugin) string {
	if strings.TrimSpace(p.CommandPath) != "" {
		return p.CommandPath
	}
	return strings.TrimSpace(p.CommandName)
}

func pluginsByCommandKey(list []plugin.Plugin) map[string]plugin.Plugin {
	m := make(map[string]plugin.Plugin, len(list))
	for _, p := range list {
		k := pluginCommandKey(p)
		if k == "" {
			continue
		}
		m[k] = p
	}
	return m
}

func diffRemovedKeys(before, after map[string]plugin.Plugin) []string {
	var keys []string
	for k := range before {
		if _, ok := after[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func emitPluginSyncDiff(
	ctx context.Context,
	log ports.Logger,
	before map[string]plugin.Plugin,
	afterList []plugin.Plugin,
	removedKeys []string,
	noRemove bool,
) SyncReport {
	var r SyncReport
	for _, p := range afterList {
		k := pluginCommandKey(p)
		if k == "" {
			continue
		}
		prev, had := before[k]
		if !had {
			r.Added++
			if log != nil {
				_ = log.Info(ctx, "Comando %q adicionado", k)
			}
			continue
		}
		if prev.ConfigHash != p.ConfigHash {
			r.Updated++
			if log != nil {
				_ = log.Info(ctx, "Comando %q atualizado", k)
			}
		}
	}
	r.Removed = len(removedKeys)
	r.AnyChange = r.Added > 0 || r.Updated > 0 || r.Removed > 0
	for _, k := range removedKeys {
		if log == nil {
			continue
		}
		if noRemove {
			_ = log.Warn(ctx, "Comando %q removido do pacote; mantido no cache (--no-remove)", k)
		} else {
			_ = log.Warn(ctx, "Comando %q deixou de existir no pacote (removido do cache)", k)
		}
	}
	return r
}

func normalizeCategoryGroupIDs(
	categories []plugin.Category,
	valid map[string]struct{},
	debug func(string),
) {
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
				debug(
					fmt.Sprintf(
						"plugin help: category_path=%q group_id=%q não cadastrado em nenhum groups.yaml; usando COMANDOS",
						c.Path,
						c.GroupID,
					),
				)
			}
			c.GroupID = ""
		}
	}
}

func normalizePluginGroupIDs(
	pluginsList []plugin.Plugin,
	valid map[string]struct{},
	debug func(string),
) {
	for i := range pluginsList {
		p := &pluginsList[i]
		if !strings.Contains(p.CommandPath, "/") {
			p.GroupID = ""
			continue
		}
		if p.GroupID == "" {
			continue
		}
		if _, ok := valid[p.GroupID]; !ok {
			if debug != nil {
				debug(
					fmt.Sprintf(
						"plugin help: command_path=%q group_id=%q não cadastrado em nenhum groups.yaml; usando COMANDOS",
						p.CommandPath,
						p.GroupID,
					),
				)
			}
			p.GroupID = ""
		}
	}
}

func checkPluginPathCollisions(pluginsList []plugin.Plugin) error {
	seen := make(map[string]string)
	for _, p := range pluginsList {
		key := p.CommandPath
		if key == "" {
			key = p.CommandName
		}
		if prevDir, ok := seen[key]; ok {
			if prevDir != p.PluginDir {
				return fmt.Errorf(
					"conflito de plugins: o caminho de commando %q está definido em dois pacotes (%s e %s). Remova ou ajuste uma das fontes",
					key,
					prevDir,
					p.PluginDir,
				)
			}
			continue
		}
		seen[key] = p.PluginDir
	}
	return nil
}

// logDebugOrNil safely calls log.Debug or does nothing if log is nil.
func logDebugOrNil(log ports.Logger, ctx context.Context, msg string) error {
	if log == nil {
		return nil
	}
	return log.Debug(ctx, "%s", msg)
}
