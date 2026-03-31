// Package update_test holds integration tests that build the full CLI root (e.g. nested
// mb tools --update-all during mb update).
package update_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cliroot "mb/internal/cli/root"
	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

func TestUpdateOnlyToolsRunsNestedToolsUpdateAll(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	toolsDir := filepath.Join(pluginsDir, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	uPath := filepath.Join(toolsDir, "u.sh")
	if err := os.WriteFile(uPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write u.sh: %v", err)
	}

	flagsJSON, err := json.Marshal(map[string]plugins.FlagDef{
		"update-all": {Type: "long", Entrypoint: "u.sh", Description: "update"},
	})
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools",
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("upsert tools: %v", err)
	}

	cfgDir := filepath.Join(tmp, "mb")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
	)
	rootCmd := cliroot.NewRootCmd(d)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetArgs([]string{"update", "--only-tools"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute update --only-tools: %v", err)
	}
}
