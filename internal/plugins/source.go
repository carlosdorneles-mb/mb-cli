package plugins

import (
	"path/filepath"
	"strings"

	"mb/internal/infra/sqlite"
)

// FirstPathSegment returns the first segment of path (before the first "/"), or path if no "/".
func FirstPathSegment(path string) string {
	if path == "" {
		return ""
	}
	idx := strings.Index(path, "/")
	if idx == -1 {
		return path
	}
	return path[:idx]
}

// PluginDirUnderRoot reports whether dir is root or a subdirectory of root.
func PluginDirUnderRoot(root, dir string) bool {
	if root == "" || dir == "" {
		return false
	}
	rel, err := filepath.Rel(root, dir)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// SourceForPlugin finds the plugin_sources row whose install root contains the plugin directory.
// Prefers the longest matching LocalPath or clone root when multiple match.
func SourceForPlugin(
	p sqlite.Plugin,
	sources []sqlite.PluginSource,
	pluginsDir string,
) *sqlite.PluginSource {
	if p.PluginDir != "" {
		var best *sqlite.PluginSource
		bestLen := -1
		for i := range sources {
			s := &sources[i]
			var root string
			if s.LocalPath != "" {
				root = s.LocalPath
			} else {
				root = filepath.Join(pluginsDir, s.InstallDir)
			}
			if PluginDirUnderRoot(root, p.PluginDir) && len(root) > bestLen {
				best = s
				bestLen = len(root)
			}
		}
		if best != nil {
			return best
		}
	}
	key := FirstPathSegment(p.CommandPath)
	for i := range sources {
		if sources[i].InstallDir == key {
			return &sources[i]
		}
	}
	return nil
}
