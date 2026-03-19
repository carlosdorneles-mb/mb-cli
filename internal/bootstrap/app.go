package bootstrap

import (
	"go.uber.org/fx"

	"mb/internal/cli/root"
	"mb/internal/module/cache"
	"mb/internal/module/cli"
	"mb/internal/module/deps"
	"mb/internal/module/executor"
	"mb/internal/module/plugins"
	runtimemod "mb/internal/module/runtime"
)

// Bootstrap builds the FX application and populates the root Cobra command.
func Bootstrap() (*fx.App, root.RootCommand, error) {
	var rootCmd root.RootCommand

	application := fx.New(
		fx.NopLogger,
		fx.Options(
			runtimemod.PathsModule,
			cache.CacheModule,
			plugins.PluginsModule,
			executor.ExecutorModule,
			deps.DepsModule,
			cli.CLIModule,
		),
		fx.Populate(&rootCmd),
	)

	if application.Err() != nil {
		return nil, nil, application.Err()
	}

	return application, rootCmd, nil
}
