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
	)
}

func TestNewUpdateCmd(t *testing.T) {
	d := testUpdateDeps(t)
	cmd := NewUpdateCmd(d)
	if cmd.Use != "update" {
		t.Errorf("Use = %q, want update", cmd.Use)
	}
	// GroupID "commands" é atribuído em internal/cli/root/command.go ao registar no root.
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if fp := cmd.Flags().Lookup("only-plugins"); fp == nil {
		t.Error("flag only-plugins missing")
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

func TestResolveUpdatePhases(t *testing.T) {
	tests := []struct {
		name                  string
		op, oc, os            bool
		wantP, wantC, wantSys bool
	}{
		{"none", false, false, false, true, true, true},
		{"only plugins", true, false, false, true, false, false},
		{"only cli", false, true, false, false, true, false},
		{"only system", false, false, true, false, false, true},
		{"plugins+cli", true, true, false, true, true, false},
		{"all three", true, true, true, true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, c, s := resolveUpdatePhases(tt.op, tt.oc, tt.os)
			if p != tt.wantP || c != tt.wantC || s != tt.wantSys {
				t.Fatalf("got plugins=%v cli=%v system=%v want plugins=%v cli=%v system=%v",
					p, c, s, tt.wantP, tt.wantC, tt.wantSys)
			}
		})
	}
}
