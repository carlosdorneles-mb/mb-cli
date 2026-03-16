package config

import (
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
