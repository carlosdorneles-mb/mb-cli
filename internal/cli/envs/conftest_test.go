package envs

// Fixtures compartilhados para testes do pacote envs (deps temporários, dirs, cache).

import (
	"os"
	"path/filepath"
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	infrakeyring "mb/internal/infra/keyring"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

func testDeps(t *testing.T) deps.Dependencies {
	t.Helper()
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	configDir := filepath.Join(tmp, "config")
	defaultEnv := filepath.Join(configDir, "env.defaults")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir plugins: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			PluginsDir:     pluginsDir,
			ConfigDir:      configDir,
			DefaultEnvPath: defaultEnv,
		},
	}
	return deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		infrakeyring.SystemKeyring{},
	)
}
