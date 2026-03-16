package app

import (
	"context"

	"go.uber.org/fx"

	"mb/internal/cache"
	"mb/internal/commands"
	"mb/internal/commands/config"
	"mb/internal/executor"
	"mb/internal/plugins"
)

func Bootstrap() (*fx.App, commands.RootCommand, error) {
	var root commands.RootCommand

	application := fx.New(
		fx.NopLogger,
		fx.Provide(
			NewAppContext,
			NewRuntimeConfig,
			NewStoreFromContext,
			NewScannerFromContext,
			executor.New,
			config.NewDependencies,
			commands.NewRootCmd,
		),
		fx.Invoke(registerLifecycle),
		fx.Populate(&root),
	)

	if application.Err() != nil {
		return nil, nil, application.Err()
	}

	return application, root, nil
}

func NewRuntimeConfig(ctx *AppContext) *config.RuntimeConfig {
	return &config.RuntimeConfig{
		ConfigDir:      ctx.ConfigDir,
		PluginsDir:     ctx.PluginsDir,
		CacheDBPath:    ctx.CacheDBPath,
		DefaultEnvPath: ctx.DefaultEnvPath,
	}
}

func NewStoreFromContext(ctx *AppContext) (*cache.Store, error) {
	return cache.NewStore(ctx.CacheDBPath)
}

func NewScannerFromContext(ctx *AppContext) *plugins.Scanner {
	return plugins.NewScanner(ctx.PluginsDir)
}

func registerLifecycle(lifecycle fx.Lifecycle, store *cache.Store) {
	lifecycle.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return store.Close()
		},
	})
}
