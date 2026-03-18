package deps

import (
	"os"
	"path/filepath"
)

// Paths holds resolved filesystem locations for the MB CLI data directory.
type Paths struct {
	ConfigDir      string
	PluginsDir     string
	CacheDBPath    string
	DefaultEnvPath string
}

// NewPaths resolves default paths under the user config dir (e.g. ~/.config/mb).
func NewPaths() (*Paths, error) {
	configBase, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(configBase, "mb")
	return &Paths{
		ConfigDir:      configDir,
		PluginsDir:     filepath.Join(configDir, "plugins"),
		CacheDBPath:    filepath.Join(configDir, "cache.db"),
		DefaultEnvPath: filepath.Join(configDir, "env.defaults"),
	}, nil
}
