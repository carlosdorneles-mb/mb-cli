package bootstrap

import (
	"io"
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"mb/internal/cli/root"
	"mb/internal/module/cache"
	"mb/internal/module/cli"
	"mb/internal/module/deps"
	"mb/internal/module/executor"
	"mb/internal/module/infra"
	"mb/internal/module/plugins"
	runtimemod "mb/internal/module/runtime"
)

// Bootstrap builds the FX application and populates the root Cobra command.
// When verbose is true, FX logs lifecycle events to stderr.
func Bootstrap(verbose bool) (*fx.App, root.RootCommand, error) {
	var rootCmd root.RootCommand

	opts := []fx.Option{
		fx.Options(
			runtimemod.PathsModule,
			cache.CacheModule,
			plugins.PluginsModule,
			executor.ExecutorModule,
			deps.DepsModule,
			infra.InfraModule,
			cli.CLIModule,
		),
		fx.Populate(&rootCmd),
	}

	if verbose {
		opts = append([]fx.Option{
			fx.WithLogger(func() fxevent.Logger {
				return &fxevent.ConsoleLogger{W: os.Stderr}
			}),
		}, opts...)
	} else {
		opts = append([]fx.Option{fx.NopLogger}, opts...)
	}

	application := fx.New(opts...)

	if application.Err() != nil {
		return nil, nil, application.Err()
	}

	return application, rootCmd, nil
}

// BootstrapWithOutput builds the FX application with a custom output writer for logs.
// Useful for testing.
func BootstrapWithOutput(w io.Writer, verbose bool) (*fx.App, root.RootCommand, error) {
	var rootCmd root.RootCommand

	opts := []fx.Option{
		fx.Options(
			runtimemod.PathsModule,
			cache.CacheModule,
			plugins.PluginsModule,
			executor.ExecutorModule,
			deps.DepsModule,
			infra.InfraModule,
			cli.CLIModule,
		),
		fx.Populate(&rootCmd),
	}

	if verbose {
		opts = append([]fx.Option{
			fx.WithLogger(func() fxevent.Logger {
				return &fxevent.ConsoleLogger{W: w}
			}),
		}, opts...)
	} else {
		opts = append([]fx.Option{fx.NopLogger}, opts...)
	}

	application := fx.New(opts...)

	if application.Err() != nil {
		return nil, nil, application.Err()
	}

	return application, rootCmd, nil
}
