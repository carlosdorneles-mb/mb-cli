package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/domain/plugin"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/ports"
	"mb/internal/shared/config"
	usecaseplugins "mb/internal/usecase/plugins"
)

func testUpdateDeps(t *testing.T) deps.Dependencies {
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

func testUpdatePluginService(t *testing.T) *usecaseplugins.UpdateService {
	t.Helper()
	d := testUpdateDeps(t)
	rt := usecaseplugins.PluginRuntime{
		ConfigDir:  d.Runtime.ConfigDir,
		PluginsDir: d.Runtime.PluginsDir,
	}
	syncSvc := usecaseplugins.NewSyncService(rt, d.Store, d.Scanner, &dummyShellInstaller{})
	return usecaseplugins.NewUpdateService(
		rt,
		d.Store,
		d.Scanner,
		&dummyShellInstaller{},
		&dummyGitOps{},
		&dummyFS{},
		syncSvc,
	)
}

type dummyShellInstaller struct{}

func (dummyShellInstaller) EnsureShellHelpers(string) (string, error) { return "", nil }

type dummyGitOps struct{}

func (dummyGitOps) ParseGitURL(
	string,
) (string, string, error) {
	return "", "", nil
}
func (dummyGitOps) Clone(context.Context, string, string, ports.GitCloneOpts) error { return nil }

func (dummyGitOps) LatestTag(
	context.Context,
	string,
) (string, error) {
	return "", nil
}

func (dummyGitOps) GetVersion(
	string,
) (string, error) {
	return "1.0.0", nil
}

func (dummyGitOps) GetCurrentBranch(
	string,
) (string, error) {
	return "main", nil
}
func (dummyGitOps) IsGitRepo(string) bool                   { return false }
func (dummyGitOps) FetchTags(context.Context, string) error { return nil }

func (dummyGitOps) ListLocalTags(
	string,
) ([]string, error) {
	return nil, nil
}

func (dummyGitOps) NewerTag(
	string,
	string,
) (string, bool) {
	return "", false
}
func (dummyGitOps) CheckoutTag(context.Context, string, string) error  { return nil }
func (dummyGitOps) FetchAndPull(context.Context, string, string) error { return nil }

type dummyFS struct{}

func (dummyFS) RemoveAll(string) error                     { return nil }
func (dummyFS) MkdirAll(string, os.FileMode) error         { return nil }
func (dummyFS) Stat(name string) (os.FileInfo, error)      { return os.Stat(name) }
func (dummyFS) IsNotExist(err error) bool                  { return os.IsNotExist(err) }
func (dummyFS) ReadDir(name string) ([]os.DirEntry, error) { return os.ReadDir(name) }
func (dummyFS) Getwd() (string, error)                     { return os.Getwd() }

func TestNewUpdateCmd(t *testing.T) {
	d := testUpdateDeps(t)
	upSvc := testUpdatePluginService(t)
	cmd := NewUpdateCmd(upSvc, d)
	if cmd.Use != "update" {
		t.Errorf("Use = %q, want update", cmd.Use)
	}
	// GroupID "commands" is set in internal/cli/root/command.go when registering on the root.
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if fp := cmd.Flags().Lookup("only-plugins"); fp == nil {
		t.Error("flag only-plugins missing")
	}
	if ft := cmd.Flags().Lookup("only-tools"); ft != nil {
		t.Error("flag only-tools should be absent without tools plugin (update-all) in cache")
	}
	if fc := cmd.Flags().Lookup("only-cli"); fc == nil {
		t.Error("flag only-cli missing")
	}
	if fs := cmd.Flags().Lookup("only-system"); fs != nil {
		t.Error("flag only-system should be absent without machine/update plugin in cache")
	}
	if fj := cmd.Flags().Lookup("json"); fj == nil {
		t.Error("flag json missing")
	}
}

func TestNewUpdateCmdOnlyToolsWithToolsPlugin(t *testing.T) {
	d := testUpdateDeps(t)
	flagsJSON, err := json.Marshal(map[string]plugins.FlagDef{
		"update-all": {Type: "long", Entrypoint: "u.sh", Description: "update"},
	})
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}
	if err := d.Store.UpsertPlugin(plugin.Plugin{
		CommandPath: toolsPluginCommandPath,
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("UpsertPlugin: %v", err)
	}
	cmd := NewUpdateCmd(testUpdatePluginService(t), d)
	if ft := cmd.Flags().Lookup("only-tools"); ft == nil {
		t.Error("flag only-tools missing when tools with update-all is in cache")
	}
}

func TestNewUpdateCmdOnlyToolsAbsentWhenToolsWithoutUpdateAll(t *testing.T) {
	d := testUpdateDeps(t)
	flagsJSON, err := json.Marshal(map[string]plugins.FlagDef{
		"other-flag": {Type: "long", Entrypoint: "x.sh", Description: "x"},
	})
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}
	if err := d.Store.UpsertPlugin(plugin.Plugin{
		CommandPath: toolsPluginCommandPath,
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("UpsertPlugin: %v", err)
	}
	cmd := NewUpdateCmd(testUpdatePluginService(t), d)
	if ft := cmd.Flags().Lookup("only-tools"); ft != nil {
		t.Error("flag only-tools should be absent when tools has no update-all flag in cache")
	}
}

func TestNewUpdateCmdOnlySystemWithMachineUpdatePlugin(t *testing.T) {
	d := testUpdateDeps(t)
	if err := d.Store.UpsertPlugin(plugin.Plugin{
		CommandPath: machineSystemUpdateCommandPath,
		CommandName: "update",
		Description: "test",
		ExecPath:    filepath.Join(t.TempDir(), "noop.sh"),
		PluginType:  "sh",
		ConfigHash:  "testhash",
	}); err != nil {
		t.Fatalf("UpsertPlugin: %v", err)
	}
	cmd := NewUpdateCmd(testUpdatePluginService(t), d)
	if fs := cmd.Flags().Lookup("only-system"); fs == nil {
		t.Error("flag only-system missing when machine/update is in cache")
	}
}

func TestUpdatePluginsAndCLICombinedNoError(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(testUpdatePluginService(t), d)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"--only-plugins", "--only-cli"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute with --only-plugins --only-cli: %v", err)
	}
}

func TestCheckOnlyWithoutOnlyCLIErrors(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(testUpdatePluginService(t), d)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"--check-only"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--check-only") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJSONWithoutOnlyCLIAndCheckOnlyErrors(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(testUpdatePluginService(t), d)
	var errBuf strings.Builder
	cmd.SetOut(os.Stdout)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--json") {
		t.Errorf("unexpected error: %v", err)
	}
}
