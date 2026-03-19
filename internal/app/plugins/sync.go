package plugins

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mb/internal/deps"
	"mb/internal/infra/plugins"
	"mb/internal/infra/shellhelpers"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/system"
)

// RunSync rescans the plugins dir and registered local paths, upserts plugins and categories, and updates the plugin_sources registry.
// Used by mb plugins sync and after plugins add/remove/update.
// log: gum log on stderr (warnings + optional success). If nil, warnings are dropped and success is not emitted.
func RunSync(
	ctx context.Context,
	d deps.Dependencies,
	log *system.Logger,
	emitSuccess bool,
) error {
	runCtx := ctx
	if runCtx == nil {
		runCtx = context.Background()
	}
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
		return err
	}
	if _, err := shellhelpers.EnsureShellHelpers(d.Runtime.ConfigDir); err != nil {
		return err
	}
	sources, err := d.Store.ListPluginSources()
	if err != nil {
		return err
	}
	for _, src := range sources {
		if src.LocalPath == "" {
			continue
		}
		p, c, w, hg, err := d.Scanner.ScanDir(src.LocalPath, src.InstallDir)
		if err != nil {
			return err
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

	if err := checkPluginPathCollisions(pluginsList); err != nil {
		return err
	}

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
		return err
	}
	if err := d.Store.DeleteAllPluginHelpGroups(); err != nil {
		return err
	}
	for _, g := range mergedHelp {
		if err := d.Store.UpsertPluginHelpGroup(
			sqlite.PluginHelpGroup{GroupID: g.ID, Title: g.Title},
		); err != nil {
			return err
		}
	}
	for _, plugin := range pluginsList {
		if err := d.Store.UpsertPlugin(plugin); err != nil {
			return err
		}
	}

	if err := d.Store.DeleteAllCategories(); err != nil {
		return err
	}
	for _, cat := range categories {
		if err := d.Store.UpsertCategory(cat); err != nil {
			return err
		}
	}

	if emitSuccess && log != nil {
		_ = log.Info(runCtx, "%d plugin(s) foram sincronizados", len(pluginsList))
	}
	return nil
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
