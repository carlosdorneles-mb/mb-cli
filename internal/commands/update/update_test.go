package update

import (
	"os"
	"path/filepath"
	"testing"

	"mb/internal/deps"
	"mb/internal/executor"
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
	)
}

func TestNewUpdateCmd(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(d)
	if cmd.Use != "update" {
		t.Errorf("Use = %q, want update", cmd.Use)
	}
	if cmd.GroupID != "commands" {
		t.Errorf("GroupID = %q, want commands", cmd.GroupID)
	}
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if fp := cmd.Flags().Lookup("only-plugins"); fp == nil {
		t.Error("flag only-plugins missing")
	}
	if fc := cmd.Flags().Lookup("only-cli"); fc == nil {
		t.Error("flag only-cli missing")
	}
}

func TestUpdateRunEBothFlagsErrors(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(d)
	cmd.SetArgs([]string{"--only-plugins", "--only-cli"})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute with both flags should return error")
	}
	if err.Error() != "não use --only-plugins e --only-cli em simultâneo" {
		t.Errorf("unexpected error: %v", err)
	}
}
