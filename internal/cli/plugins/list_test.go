package plugins

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
	"mb/internal/shared/system"
)

func TestListShowsLocalAndPath(t *testing.T) {
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "mylocal/hello",
		CommandName: "hello",
		Description: "Local plugin",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "abc",
	}); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "mylocal",
		LocalPath:  "/home/user/my-plugin",
	}); err != nil {
		t.Fatalf("upsert plugin source: %v", err)
	}

	rt := &deps.RuntimeConfig{Paths: deps.Paths{PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)
	listCmd := newPluginsListCmd(d)
	var out bytes.Buffer
	listCmd.SetOut(&out)
	listCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	// Should show simplified columns
	if !strings.Contains(out.String(), "PACOTE") {
		t.Errorf("list output should contain 'PACOTE', got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "COMANDO") {
		t.Errorf("list output should contain 'COMANDO', got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "ORIGEM") {
		t.Errorf("list output should contain 'ORIGEM', got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "local") {
		t.Errorf("list output should contain 'local', got:\n%s", out.String())
	}
	// Local path should NOT appear in simplified view (only in JSON mode)
	if strings.Contains(out.String(), "/home/user/") {
		t.Errorf(
			"list output should NOT contain full local path in simplified view, got:\n%s",
			out.String(),
		)
	}
}

func TestListEmptyRegistry(t *testing.T) {
	d := testPluginsDeps(t)
	listCmd := newPluginsListCmd(d)
	var out bytes.Buffer
	listCmd.SetOut(&out)
	listCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	// GumTable still renders headers / empty table
	if out.Len() == 0 {
		t.Error("expected some output from list")
	}
}

func TestListJSONMode(t *testing.T) {
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "test/hello",
		CommandName: "hello",
		Description: "Test plugin",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "abc",
	}); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "test",
		GitURL:     "https://github.com/org/test",
		Version:    "v1.0.0",
	}); err != nil {
		t.Fatalf("upsert plugin source: %v", err)
	}

	rt := &deps.RuntimeConfig{Paths: deps.Paths{PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)

	listCmd := newPluginsListCmd(d)
	var out bytes.Buffer
	listCmd.SetOut(&out)
	listCmd.SetErr(os.NewFile(0, os.DevNull))
	// Set --json flag
	if err := listCmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("set json flag: %v", err)
	}
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}

	// Parse JSON output
	var result struct {
		Plugins []system.PluginEntry `json:"plugins"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("parse JSON: %v\nOutput:\n%s", err, out.String())
	}

	// Verify structure
	if len(result.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(result.Plugins))
	}

	plugin := result.Plugins[0]
	if plugin.Package != "test" {
		t.Errorf("expected package 'test', got %q", plugin.Package)
	}
	if plugin.Command != "test/hello" {
		t.Errorf("expected command 'test/hello', got %q", plugin.Command)
	}
	if plugin.Description != "Test plugin" {
		t.Errorf("expected description 'Test plugin', got %q", plugin.Description)
	}
	if plugin.Version != "v1.0.0" {
		t.Errorf("expected version 'v1.0.0', got %q", plugin.Version)
	}
	if plugin.Origin != "remoto" {
		t.Errorf("expected origin 'remoto', got %q", plugin.Origin)
	}
	if plugin.URL != "https://github.com/org/test" {
		t.Errorf("expected URL, got %q", plugin.URL)
	}
}
