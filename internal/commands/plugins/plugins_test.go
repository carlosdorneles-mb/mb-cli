package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/deps"
	"mb/internal/executor"
	"mb/internal/plugins"
	"mb/internal/shared/config"
)

// testPluginsDeps returns dependencies with isolated temp dirs (cache, plugins, config).
func testPluginsDeps(t *testing.T) deps.Dependencies {
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
	store, err := cache.NewStore(cachePath)
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
	)
}

func writeMinimalRunnablePlugin(t *testing.T, dir string) {
	writeMinimalRunnablePluginNamed(t, dir, "ptest")
}

func writeMinimalRunnablePluginNamed(t *testing.T, dir, command string) {
	t.Helper()
	manifest := fmt.Sprintf("command: %s\ndescription: test plugin\nentrypoint: run.sh\n", command)
	if err := os.WriteFile(
		filepath.Join(dir, "manifest.yaml"),
		[]byte(manifest),
		0o644,
	); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "run.sh"),
		[]byte("#!/bin/sh\necho ok\n"),
		0o755,
	); err != nil {
		t.Fatalf("run.sh: %v", err)
	}
}

func TestNewPluginsCmd(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := NewPluginsCmd(d)

	if cmd.Use != "plugins" {
		t.Errorf("Use = %q, want plugins", cmd.Use)
	}
	wantAliases := []string{"plugin", "p", "extensions", "e"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Fatalf("Aliases = %v", cmd.Aliases)
	}
	for i, a := range wantAliases {
		if cmd.Aliases[i] != a {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], a)
		}
	}

	names := make(map[string]bool)
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"add", "list", "remove", "update", "sync"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestPluginsSyncRuns(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := NewPluginsCmd(d)
	var syncCmd *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Name() == "sync" {
			syncCmd = c
			break
		}
	}
	if syncCmd == nil {
		t.Fatal("sync subcommand not found")
	}
	syncCmd.SetArgs(nil)
	syncCmd.SetOut(os.Stdout)
	syncCmd.SetErr(os.Stderr)
	if err := syncCmd.Execute(); err != nil {
		t.Errorf("plugins sync Execute: %v", err)
	}
}
