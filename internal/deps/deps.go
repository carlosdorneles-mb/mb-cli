package deps

import (
	"time"

	"mb/internal/cache"
	"mb/internal/executor"
	"mb/internal/plugins"
)

// RuntimeConfig combines resolved Paths with CLI/runtime flags.
type RuntimeConfig struct {
	Paths
	Verbose         bool
	Quiet           bool
	EnvFilePath     string
	InlineEnvValues []string
	// PluginTimeout limits how long a plugin script can run. Zero means no limit.
	PluginTimeout time.Duration
}

// Dependencies groups services injected into commands.
type Dependencies struct {
	Runtime  *RuntimeConfig
	Store    *cache.Store
	Scanner  *plugins.Scanner
	Executor *executor.Executor
}

// NewDependencies constructs the dependency bundle for Fx / tests.
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
