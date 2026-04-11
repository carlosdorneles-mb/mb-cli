// Package syncdiff provides shared plugin sync logic used by both
// usecase/addplugin and usecase/plugins, eliminating code duplication.
package syncdiff

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"mb/internal/domain/plugin"
)

// SyncReport summarizes plugin command changes detected during sync.
type SyncReport struct {
	Added     int
	Updated   int
	Removed   int
	AnyChange bool
}

// Logger is the minimal interface needed for sync logging.
type Logger interface {
	Info(context.Context, string, ...any) error
	Warn(context.Context, string, ...any) error
	Debug(context.Context, string, ...any) error
}

// DebugFunc is a callback for debug messages from the scanner.
type DebugFunc func(msg string)

// NormalizeFunc validates a group ID and returns true if it is acceptable.
type NormalizeFunc func(groupID string) bool

// PluginCommandKey returns the unique key for a plugin command.
func PluginCommandKey(p plugin.Plugin) string {
	if strings.TrimSpace(p.CommandPath) != "" {
		return p.CommandPath
	}
	return strings.TrimSpace(p.CommandName)
}

// PluginsByCommandKey indexes plugins by their command key.
func PluginsByCommandKey(list []plugin.Plugin) map[string]plugin.Plugin {
	m := make(map[string]plugin.Plugin, len(list))
	for _, p := range list {
		k := PluginCommandKey(p)
		if k == "" {
			continue
		}
		m[k] = p
	}
	return m
}

// DiffRemovedKeys returns keys present in before but not in after.
func DiffRemovedKeys(before, after map[string]plugin.Plugin) []string {
	var keys []string
	for k := range before {
		if _, ok := after[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// EmitDiff logs the diff between before and after plugin lists.
func EmitDiff(
	ctx context.Context,
	log Logger,
	before map[string]plugin.Plugin,
	afterList []plugin.Plugin,
	removedKeys []string,
) SyncReport {
	var r SyncReport
	for _, p := range afterList {
		k := PluginCommandKey(p)
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
		_ = log.Warn(ctx, "Comando %q deixou de existir no pacote (removido do cache)", k)
	}
	return r
}

// NormalizePluginGroupIDs clears group_id on plugins that are not in the valid set.
// Top-level plugins (no "/" in path) always have group_id cleared.
func NormalizePluginGroupIDs(
	pluginsList []plugin.Plugin,
	validGroupIDs map[string]struct{},
	debug DebugFunc,
) {
	for i := range pluginsList {
		p := &pluginsList[i]
		normalizeGroupID(p.CommandPath, &p.GroupID, validGroupIDs, debug)
	}
}

// NormalizeCategoryGroupIDs clears group_id on categories that are not in the valid set.
// Top-level categories (no "/" in path) always have group_id cleared.
func NormalizeCategoryGroupIDs(
	categories []plugin.Category,
	validGroupIDs map[string]struct{},
	debug DebugFunc,
) {
	for i := range categories {
		c := &categories[i]
		normalizeGroupID(c.Path, &c.GroupID, validGroupIDs, debug)
	}
}

func normalizeGroupID(path string, groupID *string, valid map[string]struct{}, debug DebugFunc) {
	if !strings.Contains(path, "/") {
		*groupID = ""
		return
	}
	if *groupID == "" {
		return
	}
	if _, ok := valid[*groupID]; !ok {
		if debug != nil {
			debug(
				fmt.Sprintf(
					"plugin help: command_path=%q group_id=%q não cadastrado em nenhum groups.yaml; usando COMANDOS",
					path,
					*groupID,
				),
			)
		}
		*groupID = ""
	}
}

// CheckPluginPathCollisions returns an error if two different plugin dirs expose the same command path.
func CheckPluginPathCollisions(pluginsList []plugin.Plugin) error {
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
