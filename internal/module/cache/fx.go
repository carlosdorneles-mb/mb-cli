package cache

import (
	"context"

	"go.uber.org/fx"

	"mb/internal/deps"
	"mb/internal/infra/sqlite"
	"mb/internal/ports"
)

func newStore(p *deps.Paths) (*sqlite.Store, error) {
	return sqlite.NewStore(p.CacheDBPath)
}

func registerStoreLifecycle(lc fx.Lifecycle, store *sqlite.Store) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return store.Close()
		},
	})
}

// CacheModule provides the SQLite plugin cache and closes it on shutdown.
var CacheModule = fx.Module("cache",
	fx.Provide(
		newStore,
		func(s *sqlite.Store) ports.PluginCLIStore { return s },
	),
	fx.Invoke(registerStoreLifecycle),
)
