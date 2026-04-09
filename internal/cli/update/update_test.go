package update

import (
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
	"mb/internal/shared/config"
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

func TestNewUpdateCmd(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(d)
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
	cmd := NewUpdateCmd(d)
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
	cmd := NewUpdateCmd(d)
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
	cmd := NewUpdateCmd(d)
	if fs := cmd.Flags().Lookup("only-system"); fs == nil {
		t.Error("flag only-system missing when machine/update is in cache")
	}
}

func TestUpdatePluginsAndCLICombinedNoError(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(d)
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
	cmd := NewUpdateCmd(d)
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
	cmd := NewUpdateCmd(d)
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
