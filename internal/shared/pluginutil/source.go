// Package pluginutil provides utility functions for working with plugin domain types.
package pluginutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"mb/internal/domain/plugin"
	"mb/internal/shared/safepath"
)

// ManifestEnvVaultDefault is the logical vault when env_files omits vault or --env-vault is not set.
const ManifestEnvVaultDefault = "default"

// FlagDef defines a flag that can run an entrypoint.
type FlagDef struct {
	Type        string   `json:"type"`
	Short       string   `json:"short"`
	Entrypoint  string   `json:"entrypoint"`
	Description string   `json:"description"`
	Envs        []string `json:"envs"`
}

// EnvFileEntry represents a manifest env_files entry.
type EnvFileEntry struct {
	File  string `json:"file"`
	Vault string `json:"vault"`
}

// PluginTypeFromEntrypoint returns "sh" if entrypoint ends with .sh, otherwise "bin".
func PluginTypeFromEntrypoint(entrypoint string) string {
	if strings.HasSuffix(entrypoint, ".sh") {
		return "sh"
	}
	return "bin"
}

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
	p plugin.Plugin,
	sources []plugin.PluginSource,
	pluginsDir string,
) *plugin.PluginSource {
	if p.PluginDir != "" {
		var best *plugin.PluginSource
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

// MergeManifestEnvFiles loads KEY=VALUE from manifest env_files entries whose vault
// matches effectiveVault (use ManifestEnvVaultDefault when --env-vault is unset).
// Later files in the list override earlier ones for duplicate keys.
func MergeManifestEnvFiles(
	pluginDir, envFilesJSON, effectiveVault string,
) (map[string]string, error) {
	if envFilesJSON == "" || pluginDir == "" {
		return map[string]string{}, nil
	}
	var entries []EnvFileEntry
	if err := json.Unmarshal([]byte(envFilesJSON), &entries); err != nil {
		return nil, fmt.Errorf("env_files do plugin: JSON inválido: %w", err)
	}
	merged := map[string]string{}
	for _, e := range entries {
		if e.Vault != effectiveVault {
			continue
		}
		abs := filepath.Join(pluginDir, e.File)
		if err := safepath.ValidateUnderDir(abs, pluginDir); err != nil {
			return nil, fmt.Errorf("env_files %q: path inválido: %w", e.File, err)
		}
		values, err := godotenv.Read(abs)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("arquivo de ambiente do plugin não encontrado: %s", e.File)
			}
			return nil, fmt.Errorf("ler env_files %q: %w", e.File, err)
		}
		for k, v := range values {
			merged[k] = v
		}
	}
	return merged, nil
}
