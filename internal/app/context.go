package app

import (
	"os"
	"path/filepath"
)

type AppContext struct {
	ConfigDir       string
	PluginsDir      string
	CacheDBPath     string
	DefaultEnvPath  string
	Verbose         bool
	Quiet           bool
	EnvFilePath     string
	InlineEnvValues []string
}

func NewAppContext() (*AppContext, error) {
	configBase, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(configBase, "mb")
	return &AppContext{
		ConfigDir:      configDir,
		PluginsDir:     filepath.Join(configDir, "plugins"),
		CacheDBPath:    filepath.Join(configDir, "cache.db"),
		DefaultEnvPath: filepath.Join(configDir, "env.defaults"),
	}, nil
}
