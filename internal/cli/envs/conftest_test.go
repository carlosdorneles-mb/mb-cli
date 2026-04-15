package envs

// Fixtures compartilhados para testes do pacote envs (deps temporários, dirs, cache).

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	infrakeyring "mb/internal/infra/keyring"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/ports"
	"mb/internal/shared/config"
	"mb/internal/usecase/envs"
)

// memorySecretStore is an in-memory SecretStore for tests (avoids D-Bus / org.freedesktop.secrets on CI).
type memorySecretStore map[string]string

func (m memorySecretStore) key(group, k string) string { return group + "\x00" + k }

func (m memorySecretStore) Set(group, k, v string) error {
	m[m.key(group, k)] = v
	return nil
}

func (m memorySecretStore) Get(group, k string) (string, error) {
	v, ok := m[m.key(group, k)]
	if !ok {
		return "", errors.New("missing")
	}
	return v, nil
}

func (m memorySecretStore) Delete(group, k string) error {
	delete(m, m.key(group, k))
	return nil
}

var _ ports.SecretStore = memorySecretStore{}

func testDeps(t *testing.T) deps.Dependencies {
	t.Helper()
	return testDepsWithSecretStore(t, infrakeyring.SystemKeyring{})
}

func testDepsWithSecretStore(t *testing.T, secrets ports.SecretStore) deps.Dependencies {
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
		secrets,
		nil,
	)
}

func testListServiceForDeps(t *testing.T, d deps.Dependencies) *envs.ListService {
	t.Helper()
	paths := envs.Paths{
		DefaultEnvPath: d.Runtime.DefaultEnvPath,
		ConfigDir:      d.Runtime.ConfigDir,
	}
	return envs.NewListService(d.SecretStore, d.OnePassword, paths)
}
