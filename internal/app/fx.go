package app

import (
	"context"

	"go.uber.org/fx"

	"mb/internal/commands"
	"mb/internal/deps"
	"mb/internal/infra/sqlite"
	"mb/internal/plugins"
	"mb/internal/shared/config"
)

func Bootstrap() (*fx.App, commands.RootCommand, error) {
	var root commands.RootCommand

	application := fx.New(
		fx.NopLogger,
		fx.Options(
			PathsModule,
			CacheModule,
			PluginsModule,
			ExecutorModule,
			DepsModule,
			CLIModule,
		),
		fx.Populate(&root),
	)

	if application.Err() != nil {
		return nil, nil, application.Err()
	}

	return application, root, nil
}

// NewRuntimeConfig builds runtime config from resolved paths (CLI flags stay at zero values until Cobra runs).
func NewRuntimeConfig(p *deps.Paths) *deps.RuntimeConfig {
	return &deps.RuntimeConfig{Paths: *p}
}

// NewAppConfig loads ~/.config/mb/config.yaml (Viper + precedence for known keys).
func NewAppConfig(p *deps.Paths) (config.AppConfig, error) {
	return config.Load(p.ConfigDir)
}

func newStore(p *deps.Paths) (*sqlite.Store, error) {
	return sqlite.NewStore(p.CacheDBPath)
}

func newScanner(p *deps.Paths) *plugins.Scanner {
	return plugins.NewScanner(p.PluginsDir)
}

func registerStoreLifecycle(lifecycle fx.Lifecycle, store *sqlite.Store) {
	lifecycle.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return store.Close()
		},
	})
}
