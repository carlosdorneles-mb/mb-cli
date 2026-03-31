package plugincmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/cli/root"
	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

// Repro: leaf manifest at tools/ (flags-only) plus tools/bruno with group_id from groups.yaml
// used to panic: group id 'development' is not defined for subcommand 'mb tools bruno'.
func TestLeafToolsWithNestedBrunoHelpGroupNoPanic(t *testing.T) {
	flagsJSON, err := json.Marshal(map[string]plugins.FlagDef{
		"update-all": {Type: "long", Entrypoint: "u.sh", Description: "update"},
	})
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	brunoDir := filepath.Join(pluginsDir, "tools", "bruno")
	if err := os.MkdirAll(brunoDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(brunoDir, "index.sh"),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o755,
	); err != nil {
		t.Fatalf("write index.sh: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPluginHelpGroup(sqlite.PluginHelpGroup{
		GroupID: "development",
		Title:   "Desenvolvimento",
	}); err != nil {
		t.Fatalf("upsert help group: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools",
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("upsert tools leaf: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools/bruno",
		CommandName: "bruno",
		Description: "Bruno",
		ExecPath:    filepath.Join(brunoDir, "index.sh"),
		PluginType:  "sh",
		ConfigHash:  "b1",
		GroupID:     "development",
		PluginDir:   brunoDir,
	}); err != nil {
		t.Fatalf("upsert tools/bruno: %v", err)
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
	r := root.NewRootCmd(d)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Execute panicked (Cobra group mismatch): %v", r)
		}
	}()
	r.SetArgs([]string{})
	var out strings.Builder
	r.SetOut(&out)
	r.SetErr(&out)
	if err := r.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}
