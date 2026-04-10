package plugins

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/domain/plugin"
	"mb/internal/fakes"
	"mb/internal/infra/executor"
	infrakeyring "mb/internal/infra/keyring"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/ports"
	"mb/internal/shared/config"
	"mb/internal/usecase/addplugin"
)

// Ensure the gitOpsForTestImpl satisfies ports.GitOperations.
var _ ports.GitOperations = (*gitOpsForTestImpl)(nil)

// testAddService builds an AddPlugin service backed by real SQLite and OS FS
// for integration tests that exercise the full use case.
func testAddService(t *testing.T) (*addplugin.Service, *deps.RuntimeConfig) {
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

	syncer := addplugin.NewSyncer()
	svc := addplugin.New(
		addplugin.Runtime{
			ConfigDir:  rt.ConfigDir,
			PluginsDir: rt.PluginsDir,
		},
		store,
		plugins.NewScanner(pluginsDir),
		&osFSAdapter{},
		gitOpsForTest(),
		&shellInstaller{},
		&layoutValidator{},
		syncer,
	)
	return svc, rt
}

// osFSAdapter adapts the real OS to the ports.Filesystem interface for tests.
type osFSAdapter struct{}

func (osFSAdapter) RemoveAll(path string) error                { return os.RemoveAll(path) }
func (osFSAdapter) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (osFSAdapter) Stat(name string) (os.FileInfo, error)      { return os.Stat(name) }
func (osFSAdapter) IsNotExist(err error) bool                  { return os.IsNotExist(err) }
func (osFSAdapter) ReadDir(name string) ([]os.DirEntry, error) { return os.ReadDir(name) }
func (osFSAdapter) Getwd() (string, error)                     { return os.Getwd() }

type shellInstaller struct{}

func (shellInstaller) EnsureShellHelpers(configDir string) (string, error) {
	return configDir, nil
}

type layoutValidator struct{}

func (layoutValidator) ValidatePluginRoot(dir string) error { return nil }

type gitOpsForTestImpl struct{}

func (gitOpsForTestImpl) ParseGitURL(raw string) (string, string, error) {
	// Only treat https://, git@, ssh:// as Git URLs
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "git@") || strings.HasPrefix(raw, "ssh://") {
		parts := strings.Split(raw, "/")
		name := parts[len(parts)-1]
		if strings.HasSuffix(name, ".git") {
			name = strings.TrimSuffix(name, ".git")
		}
		return name, raw, nil
	}
	// Return error so addplugin.Service falls through to addLocal
	return "", "", fmt.Errorf("not a git URL")
}

func (gitOpsForTestImpl) Clone(ctx context.Context, repoURL, destDir string, opts ports.GitCloneOpts) error {
	return nil
}

func (gitOpsForTestImpl) LatestTag(ctx context.Context, repoURL string) (string, error) {
	return "", nil
}

func (gitOpsForTestImpl) GetVersion(dir string) (string, error) { return "1.0.0", nil }

func (gitOpsForTestImpl) GetCurrentBranch(dir string) (string, error) { return "main", nil }

func (gitOpsForTestImpl) IsGitRepo(dir string) bool { return false }

func (gitOpsForTestImpl) FetchTags(ctx context.Context, dir string) error { return nil }

func (gitOpsForTestImpl) ListLocalTags(dir string) ([]string, error) { return nil, nil }

func (gitOpsForTestImpl) NewerTag(current, candidate string) (string, bool) { return "", false }

func (gitOpsForTestImpl) CheckoutTag(ctx context.Context, dir, tag string) error { return nil }

func (gitOpsForTestImpl) FetchAndPull(ctx context.Context, dir, ref string) error { return nil }

func gitOpsForTest() *gitOpsForTestImpl { return &gitOpsForTestImpl{} }

// testPluginsDepsWithAdd returns dependencies and an AddPlugin service for CLI-level tests.
// Both share the same SQLite store so assertions work correctly.
func testPluginsDepsWithAdd(t *testing.T) (deps.Dependencies, *addplugin.Service) {
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

	syncer := addplugin.NewSyncer()
	svc := addplugin.New(
		addplugin.Runtime{
			ConfigDir:  rt.ConfigDir,
			PluginsDir: rt.PluginsDir,
		},
		store,
		plugins.NewScanner(pluginsDir),
		&osFSAdapter{},
		gitOpsForTest(),
		&shellInstaller{},
		&layoutValidator{},
		syncer,
	)

	d := deps.Dependencies{
		Runtime:     rt,
		AppConfig:   config.AppConfig{},
		Store:       store,
		Scanner:     plugins.NewScanner(pluginsDir),
		Executor:    executor.New(),
		SecretStore: infrakeyring.SystemKeyring{},
		OnePassword: nil,
	}

	return d, svc
}

// --- Tests using the new AddPlugin Service ---

func TestAddRequiresExactlyOneArg(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error with no args")
	}

	cmd.SetArgs([]string{"a", "b"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error with two args")
	}
}

func TestAddLocalPathNotFound(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	cmd := newPluginsAddCmd(svc, d)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	missing := filepath.Join(t.TempDir(), "nope")
	cmd.SetArgs([]string{missing})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "não encontrado") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddLocalPathNotDirectory(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	cmd := newPluginsAddCmd(svc, d)
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "não é um diretório") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddLocalNoManifest(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	emptyDir := t.TempDir()
	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{emptyDir})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nenhum plugin encontrado") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddCollectionTwoPlugins(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	parent := t.TempDir()
	for _, pair := range []struct{ dir, cmd string }{{"alpha", "acmd"}, {"beta", "bcmd"}} {
		dir := filepath.Join(parent, pair.dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeMinimalRunnablePluginNamed(t, dir, pair.cmd)
	}
	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetArgs([]string{parent})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add collection: %v", err)
	}
	// Get the store from the service deps
	store := d.Store
	for _, name := range []string{"alpha", "beta"} {
		src, err := store.GetPluginSource(name)
		if err != nil || src == nil {
			t.Fatalf("missing source %q: %v", name, err)
		}
	}
}

func TestAddCollectionPackageWithMultipleFails(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	parent := t.TempDir()
	for _, pair := range []struct{ dir, cmd string }{{"a", "acmd"}, {"b", "bcmd"}} {
		dir := filepath.Join(parent, pair.dir)
		_ = os.MkdirAll(dir, 0o755)
		writeMinimalRunnablePluginNamed(t, dir, pair.cmd)
	}
	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{parent, "--package", "x"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--package") {
		t.Fatalf("expected --package error: %v", err)
	}
}

func TestAddCollectionSkipsInvalidSubdir(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	parent := t.TempDir()
	good := filepath.Join(parent, "good")
	_ = os.MkdirAll(good, 0o755)
	writeMinimalRunnablePlugin(t, good)
	bad := filepath.Join(parent, "bad")
	_ = os.MkdirAll(bad, 0o755)
	_ = os.WriteFile(
		filepath.Join(bad, "manifest.yaml"),
		[]byte("command: x\ndescription: y\nentrypoint: missing.sh\n"),
		0o644,
	)

	var errBuf bytes.Buffer
	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{parent})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(errBuf.String(), "bad") {
		t.Errorf("expected stderr warn about bad: %s", errBuf.String())
	}
	store := d.Store
	src, _ := store.GetPluginSource("good")
	if src == nil {
		t.Fatal("good should be registered")
	}
}

func TestAddCollectionSingleWithCustomPackage(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	parent := t.TempDir()
	dir := filepath.Join(parent, "orig")
	_ = os.MkdirAll(dir, 0o755)
	writeMinimalRunnablePlugin(t, dir)
	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetArgs([]string{parent, "--package", "custom"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	store := d.Store
	if _, err := store.GetPluginSource("custom"); err != nil {
		t.Fatal(err)
	}
}

func TestAddLocalRegistersPlugin(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	pluginDir := t.TempDir()
	writeMinimalRunnablePlugin(t, pluginDir)

	cmd := newPluginsAddCmd(svc, d)
	var errBuf bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{pluginDir, "--package", "myplug"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	store := d.Store
	src, err := store.GetPluginSource("myplug")
	if err != nil {
		t.Fatalf("GetPluginSource: %v", err)
	}
	if src == nil {
		t.Fatal("expected plugin source")
	}
	if src.LocalPath == "" {
		t.Error("expected LocalPath set")
	}
	if !strings.Contains(errBuf.String(), "myplug") {
		t.Errorf("log should mention package id: %s", errBuf.String())
	}
}

func TestAddLocalReplaceExistingPackage(t *testing.T) {
	d, svc := testPluginsDepsWithAdd(t)
	store := d.Store
	if err := store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "taken",
		LocalPath:  "/tmp/x",
	}); err != nil {
		t.Fatal(err)
	}

	pluginDir := t.TempDir()
	writeMinimalRunnablePlugin(t, pluginDir)

	cmd := newPluginsAddCmd(svc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{pluginDir, "--package", "taken"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("add replace: %v", err)
	}
	src, err := store.GetPluginSource("taken")
	if err != nil || src == nil {
		t.Fatal("expected plugin source")
	}
	if src.LocalPath != pluginDir {
		t.Errorf("LocalPath = %q, want %q", src.LocalPath, pluginDir)
	}
}

// --- Unit tests for the AddPlugin Service using fakes ---

func TestAddPluginService_InvalidPath(t *testing.T) {
	fsys := fakes.NewFakeFS()
	git := fakes.NewFakeGit()
	shell := &fakes.FakeShellInstaller{}
	layout := &fakes.FakeLayoutValidator{}
	logger := fakes.NewFakeLogger()

	store := &fakeStoreForAdd{sources: make(map[string]plugin.PluginSource)}
	scanner := &fakes.FakePluginScanner{}
	syncer := addplugin.NewSyncer()

	rt := addplugin.Runtime{
		ConfigDir:  "/tmp/config",
		PluginsDir: "/tmp/plugins",
	}
	svc := addplugin.New(rt, store, scanner, fsys, git, shell, layout, syncer)

	// Use a path that definitely doesn't exist in the fake FS
	err := svc.Add(t.Context(), addplugin.Request{Source: "/path/does/not/exist"}, logger)
	t.Logf("err = %v", err)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), "não encontrado") && !strings.Contains(err.Error(), "acesso ao diretório") {
		t.Errorf("unexpected error: %v", err)
	}
}

// fakeStoreForAdd implements ports.PluginCacheStore for unit tests.
type fakeStoreForAdd struct {
	sources  map[string]plugin.PluginSource
	plugins  []plugin.Plugin
	categories []plugin.Category
	helpGroups []plugin.PluginHelpGroup
}

func (f *fakeStoreForAdd) GetPluginSource(id string) (*plugin.PluginSource, error) {
	if s, ok := f.sources[id]; ok {
		return &s, nil
	}
	return nil, nil
}

func (f *fakeStoreForAdd) UpsertPluginSource(s plugin.PluginSource) error {
	src := s
	f.sources[s.InstallDir] = src
	return nil
}

func (f *fakeStoreForAdd) ListPluginSources() ([]plugin.PluginSource, error) {
	result := make([]plugin.PluginSource, 0, len(f.sources))
	for _, s := range f.sources {
		result = append(result, s)
	}
	return result, nil
}

func (f *fakeStoreForAdd) DeletePluginSource(id string) error {
	delete(f.sources, id)
	return nil
}

func (f *fakeStoreForAdd) DeleteAllPluginSources() error {
	f.sources = make(map[string]plugin.PluginSource)
	return nil
}

// PluginSyncStore methods
func (f *fakeStoreForAdd) ListPlugins() ([]plugin.Plugin, error) {
	return f.plugins, nil
}

func (f *fakeStoreForAdd) DeleteAllPlugins() error {
	f.plugins = nil
	return nil
}

func (f *fakeStoreForAdd) DeleteAllPluginHelpGroups() error {
	f.helpGroups = nil
	return nil
}

func (f *fakeStoreForAdd) UpsertPluginHelpGroup(g plugin.PluginHelpGroup) error {
	f.helpGroups = append(f.helpGroups, g)
	return nil
}

func (f *fakeStoreForAdd) UpsertPlugin(p plugin.Plugin) error {
	f.plugins = append(f.plugins, p)
	return nil
}

func (f *fakeStoreForAdd) DeleteAllCategories() error {
	f.categories = nil
	return nil
}

func (f *fakeStoreForAdd) UpsertCategory(c plugin.Category) error {
	f.categories = append(f.categories, c)
	return nil
}

// PluginCLIStore methods
func (f *fakeStoreForAdd) ListCategories() ([]plugin.Category, error) {
	return f.categories, nil
}

func (f *fakeStoreForAdd) ListPluginHelpGroups() ([]plugin.PluginHelpGroup, error) {
	return f.helpGroups, nil
}
