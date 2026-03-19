package plugins

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/cache"
	"mb/internal/config"
	"mb/internal/deps"
	"mb/internal/executor"
	"mb/internal/plugins"
)

func TestListShowsLocalAndPath(t *testing.T) {
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "mylocal/hello",
		CommandName: "hello",
		Description: "Local plugin",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "abc",
	}); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPluginSource(cache.PluginSource{
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
	)
	listCmd := newPluginsListCmd(d)
	var out bytes.Buffer
	listCmd.SetOut(&out)
	listCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out.String(), "local") {
		t.Errorf("list output should contain 'local', got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "/home/user/my-plugin") {
		t.Errorf("list output should contain local path, got:\n%s", out.String())
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
