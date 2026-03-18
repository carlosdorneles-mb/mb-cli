package app

import (
	"go.uber.org/fx"

	"mb/internal/commands"
	"mb/internal/deps"
	"mb/internal/executor"
)

// PathsModule resolves MB data directory paths and runtime shell (paths only until flags parse).
var PathsModule = fx.Module("paths",
	fx.Provide(deps.NewPaths),
	fx.Provide(NewRuntimeConfig),
)

// DepsModule bundles injected services for commands.
var DepsModule = fx.Module("deps",
	fx.Provide(deps.NewDependencies),
)

// CacheModule provides the SQLite plugin cache and closes it on shutdown.
var CacheModule = fx.Module("cache",
	fx.Provide(newStore),
	fx.Invoke(registerStoreLifecycle),
)

// PluginsModule provides the plugin filesystem scanner.
var PluginsModule = fx.Module("plugins",
	fx.Provide(newScanner),
)

// ExecutorModule provides the plugin script runner.
var ExecutorModule = fx.Module("executor",
	fx.Provide(executor.New),
)

// CLIModule wires the root Cobra command.
var CLIModule = fx.Module("cli",
	fx.Provide(commands.NewRootCmd),
)
