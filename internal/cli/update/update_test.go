package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
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
	if ft := cmd.Flags().Lookup("only-tools"); ft == nil {
		t.Error("flag only-tools missing")
	}
	if fc := cmd.Flags().Lookup("only-cli"); fc == nil {
		t.Error("flag only-cli missing")
	}
	if fs := cmd.Flags().Lookup("only-system"); fs == nil {
		t.Error("flag only-system missing")
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
