package root

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/ports"
	"mb/internal/shared/config"
)

// testRootDeps builds Dependencies with an isolated temp SQLite store.
func testRootDeps(t *testing.T) deps.Dependencies {
	t.Helper()
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	configDir := filepath.Join(tmp, "config")
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
		nil,
		nil,
	)
}

// testRootInfraPorts returns the four infra interfaces needed by NewRootCmd.
func testRootInfraPorts(t *testing.T) (ports.Filesystem, ports.GitOperations, ports.ShellHelperInstaller, ports.PluginLayoutValidator) {
	t.Helper()
	return &testOSFS{}, &testGitOps{}, &testShellInstaller{}, &testLayoutValidator{}
}

// Minimal test adapters
type testOSFS struct{}

func (testOSFS) RemoveAll(string) error                     { return nil }
func (testOSFS) MkdirAll(string, os.FileMode) error         { return nil }
func (testOSFS) Stat(name string) (os.FileInfo, error)      { return os.Stat(name) }
func (testOSFS) IsNotExist(err error) bool                  { return os.IsNotExist(err) }
func (testOSFS) ReadDir(name string) ([]os.DirEntry, error) { return os.ReadDir(name) }
func (testOSFS) Getwd() (string, error)                     { return os.Getwd() }

type testGitOps struct{}

func (testGitOps) ParseGitURL(raw string) (string, string, error) {
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "git@") {
		return "repo", raw, nil
	}
	return "", "", nil
}
func (testGitOps) Clone(context.Context, string, string, ports.GitCloneOpts) error {
	return nil
}
func (testGitOps) LatestTag(context.Context, string) (string, error)   { return "", nil }
func (testGitOps) GetVersion(string) (string, error)                   { return "1.0.0", nil }
func (testGitOps) GetCurrentBranch(string) (string, error)             { return "main", nil }
func (testGitOps) IsGitRepo(string) bool                               { return false }
func (testGitOps) FetchTags(context.Context, string) error             { return nil }
func (testGitOps) ListLocalTags(string) ([]string, error)              { return nil, nil }
func (testGitOps) NewerTag(string, string) (string, bool)              { return "", false }
func (testGitOps) CheckoutTag(context.Context, string, string) error   { return nil }
func (testGitOps) FetchAndPull(context.Context, string, string) error  { return nil }

type testShellInstaller struct{}

func (testShellInstaller) EnsureShellHelpers(string) (string, error) { return "", nil }

type testLayoutValidator struct{}

func (testLayoutValidator) ValidatePluginRoot(string) error { return nil }
