package plugins

import (
	"context"
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
	"mb/internal/usecase/addplugin"
	usecaseplugins "mb/internal/usecase/plugins"
)

// testAllPluginServices builds all plugin usecase services for tests.
// It returns the services AND the shared deps so tests can use the same store.
func testAllPluginServicesWithDeps(t *testing.T) (*addplugin.Service, *usecaseplugins.SyncService, *usecaseplugins.RemoveService, *usecaseplugins.UpdateService, deps.Dependencies) {
	t.Helper()
	d := testPluginDeps(t)

	rt := usecaseplugins.PluginRuntime{
		ConfigDir:  d.Runtime.ConfigDir,
		PluginsDir: d.Runtime.PluginsDir,
		Quiet:      false,
		Verbose:    false,
	}

	scanner := d.Scanner
	shell := &testShellInstaller{}
	fsys := &testOSFS{}
	git := &testGitOps{}
	layout := &testLayoutValidator{}
	syncer := addplugin.NewSyncer()

	syncSvc := usecaseplugins.NewSyncService(rt, d.Store, scanner, shell)
	addSvc := addplugin.New(
		addplugin.Runtime{ConfigDir: d.Runtime.ConfigDir, PluginsDir: d.Runtime.PluginsDir},
		d.Store, scanner, fsys, git, shell, layout, syncer,
	)
	rmSvc := usecaseplugins.NewRemoveService(rt, d.Store, scanner, shell, fsys, syncSvc)
	upSvc := usecaseplugins.NewUpdateService(rt, d.Store, scanner, shell, git, fsys, syncSvc)

	return addSvc, syncSvc, rmSvc, upSvc, d
}

// testAllPluginServices builds all plugin usecase services for tests.
func testAllPluginServices(t *testing.T) (*addplugin.Service, *usecaseplugins.SyncService, *usecaseplugins.RemoveService, *usecaseplugins.UpdateService) {
	t.Helper()
	a, s, r, u, _ := testAllPluginServicesWithDeps(t)
	return a, s, r, u
}

// testPluginDeps builds Dependencies for tests that don't need the services directly.
func testPluginDeps(t *testing.T) deps.Dependencies {
	t.Helper()
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	configDir := filepath.Join(tmp, "config")
	_ = os.MkdirAll(pluginsDir, 0o755)
	_ = os.MkdirAll(configDir, 0o755)

	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			PluginsDir: pluginsDir,
			ConfigDir:  configDir,
		},
	}
	return deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		infrakeyring.SystemKeyring{},
		nil,
	)
}

type testOSFS struct{}

func (testOSFS) RemoveAll(string) error                          { return nil }
func (testOSFS) MkdirAll(string, os.FileMode) error              { return nil }
func (testOSFS) Stat(name string) (os.FileInfo, error)           { return os.Stat(name) }
func (testOSFS) IsNotExist(err error) bool                       { return os.IsNotExist(err) }
func (testOSFS) ReadDir(name string) ([]os.DirEntry, error)      { return os.ReadDir(name) }
func (testOSFS) Getwd() (string, error)                          { return os.Getwd() }

type testGitOps struct{}

func (testGitOps) ParseGitURL(string) (string, string, error)    { return "", "", nil }
func (testGitOps) Clone(context.Context, string, string, ports.GitCloneOpts) error { return nil }
func (testGitOps) LatestTag(context.Context, string) (string, error) { return "", nil }
func (testGitOps) GetVersion(string) (string, error)             { return "1.0.0", nil }
func (testGitOps) GetCurrentBranch(string) (string, error)       { return "main", nil }
func (testGitOps) IsGitRepo(string) bool                         { return false }
func (testGitOps) FetchTags(context.Context, string) error       { return nil }
func (testGitOps) ListLocalTags(string) ([]string, error)        { return nil, nil }
func (testGitOps) NewerTag(string, string) (string, bool)        { return "", false }
func (testGitOps) CheckoutTag(context.Context, string, string) error { return nil }
func (testGitOps) FetchAndPull(context.Context, string, string) error { return nil }

type testShellInstaller struct{}

func (testShellInstaller) EnsureShellHelpers(string) (string, error) { return "", nil }

type testLayoutValidator struct{}

func (testLayoutValidator) ValidatePluginRoot(string) error      { return nil }
