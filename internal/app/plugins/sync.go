package plugins

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"mb/internal/deps"
	"mb/internal/infra/plugins"
	"mb/internal/infra/shellhelpers"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/system"
)

// SyncOptions configures RunSync behaviour.
type SyncOptions struct {
	// EmitSuccess logs a short summary when true (e.g. plugins sync command): one line when
	// nothing changed, or no extra bulk line when per-command diff lines were emitted.
	EmitSuccess bool
	// NoRemove keeps SQLite rows for commands that disappeared from scanned packages (orphans).
	NoRemove bool
}

// SyncReport summarizes plugin command changes detected during sync (leaf commands only).
type SyncReport struct {
	Added     int
	Updated   int
	Removed   int
	AnyChange bool
}

// RunSync rescans the plugins dir and registered local paths, upserts plugins and categories, and updates the plugin_sources registry.
// Used by mb plugins sync and after plugins add/remove/update.
// log: gum log on stderr (warnings + optional success). If nil, warnings are dropped and success is not emitted.
func RunSync(
	ctx context.Context,
	d deps.Dependencies,
	log *system.Logger,
	opts SyncOptions,
) (SyncReport, error) {
	runCtx := ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	beforePlugins, err := d.Store.ListPlugins()
	if err != nil {
		return SyncReport{}, err
	}
	beforeByKey := pluginsByCommandKey(beforePlugins)

	pluginHelpLog := log
	if pluginHelpLog == nil {
		quiet, verbose := false, false
		if d.Runtime != nil {
			quiet, verbose = d.Runtime.Quiet, d.Runtime.Verbose
		}
		pluginHelpLog = system.NewLogger(quiet, verbose, os.Stderr)
	}
	d.Scanner.DebugLog = func(msg string) { _ = pluginHelpLog.Debug(runCtx, "%s", msg) }
	defer func() { d.Scanner.DebugLog = nil }()

	pluginsList, categories, warnings, hgBatches, err := d.Scanner.Scan()
	if err != nil {
		return SyncReport{}, err
	}
	if _, err := shellhelpers.EnsureShellHelpers(d.Runtime.ConfigDir); err != nil {
		return SyncReport{}, err
	}
	sources, err := d.Store.ListPluginSources()
	if err != nil {
		return SyncReport{}, err
	}
	for _, src := range sources {
		if src.LocalPath == "" {
			continue
		}
		p, c, w, hg, err := d.Scanner.ScanDir(src.LocalPath, src.InstallDir)
		if err != nil {
			return SyncReport{}, err
		}
		pluginsList = append(pluginsList, p...)
		categories = append(categories, c...)
		warnings = append(warnings, w...)
		hgBatches = append(hgBatches, hg...)
	}
	mergedHelp := plugins.MergeHelpGroupsGlobal(hgBatches, func(msg string) {
		_ = pluginHelpLog.Debug(runCtx, "%s", msg)
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
		_ = pluginHelpLog.Debug(runCtx, "%s", msg)
	})
	normalizeCategoryGroupIDs(categories, validGroupIDs, func(msg string) {
		_ = pluginHelpLog.Debug(runCtx, "%s", msg)
	})

	if err := d.Store.DeleteAllPlugins(); err != nil {
		return SyncReport{}, err
	}
	if err := d.Store.DeleteAllPluginHelpGroups(); err != nil {
		return SyncReport{}, err
	}
	for _, g := range mergedHelp {
		if err := d.Store.UpsertPluginHelpGroup(
			sqlite.PluginHelpGroup{GroupID: g.ID, Title: g.Title},
		); err != nil {
			return SyncReport{}, err
		}
	}
	for _, plugin := range pluginsList {
		if err := d.Store.UpsertPlugin(plugin); err != nil {
			return SyncReport{}, err
		}
	}

	if err := d.Store.DeleteAllCategories(); err != nil {
		return SyncReport{}, err
	}
	for _, cat := range categories {
		if err := d.Store.UpsertCategory(cat); err != nil {
			return SyncReport{}, err
		}
	}

	if opts.EmitSuccess && log != nil && !report.AnyChange {
		_ = log.Info(runCtx, "Nenhum comando alterado; cache atualizado.")
	}
	return report, nil
}

func pluginCommandKey(p sqlite.Plugin) string {
	if strings.TrimSpace(p.CommandPath) != "" {
		return p.CommandPath
	}
	return strings.TrimSpace(p.CommandName)
}

func pluginsByCommandKey(list []sqlite.Plugin) map[string]sqlite.Plugin {
	m := make(map[string]sqlite.Plugin, len(list))
	for _, p := range list {
		k := pluginCommandKey(p)
		if k == "" {
			continue
		}
		m[k] = p
	}
	return m
}

func diffRemovedKeys(before, after map[string]sqlite.Plugin) []string {
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
	log *system.Logger,
	before map[string]sqlite.Plugin,
	afterList []sqlite.Plugin,
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
	categories []sqlite.Category,
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
	pluginsList []sqlite.Plugin,
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

func checkPluginPathCollisions(pluginsList []sqlite.Plugin) error {
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
