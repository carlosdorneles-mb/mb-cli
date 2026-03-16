package config

import (
	"time"

	"mb/internal/cache"
	"mb/internal/executor"
	"mb/internal/plugins"
)

type RuntimeConfig struct {
	ConfigDir       string
	PluginsDir      string
	CacheDBPath     string
	DefaultEnvPath  string
	Verbose         bool
	Quiet           bool
	EnvFilePath     string
	InlineEnvValues []string
	// PluginTimeout limits how long a plugin script can run. Zero means no limit.
	PluginTimeout time.Duration
}

type Dependencies struct {
	Runtime  *RuntimeConfig
	Store    *cache.Store
	Scanner  *plugins.Scanner
	Executor *executor.Executor
}

func NewDependencies(
	runtime *RuntimeConfig,
	store *cache.Store,
	scanner *plugins.Scanner,
	exec *executor.Executor,
) Dependencies {
	return Dependencies{
		Runtime:  runtime,
		Store:    store,
		Scanner:  scanner,
		Executor: exec,
	}
}
