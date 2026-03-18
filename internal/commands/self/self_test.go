package self

import (
	"os"
	"path/filepath"
	"testing"

	"mb/internal/cache"
	"mb/internal/deps"
	"mb/internal/executor"
	"mb/internal/plugins"
)

func testSelfDeps(t *testing.T) deps.Dependencies {
	t.Helper()
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	configDir := filepath.Join(tmp, "config")
	defaultEnv := filepath.Join(configDir, "env.defaults")
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
			PluginsDir:     pluginsDir,
			ConfigDir:      configDir,
			DefaultEnvPath: defaultEnv,
		},
	}
	return deps.NewDependencies(rt, store, plugins.NewScanner(pluginsDir), executor.New())
}

func writePluginWithCommand(t *testing.T, dir, command string) {
	t.Helper()
	m := "command: " + command + "\ndescription: x\nentrypoint: run.sh\n"
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(m), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestNewSelfCmd(t *testing.T) {
	d := testSelfDeps(t)
	cmd := NewSelfCmd(d)

	if cmd.Use != "self" {
		t.Errorf("Use = %q", cmd.Use)
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "s" {
		t.Errorf("Aliases = %v", cmd.Aliases)
	}

	want := map[string]bool{"sync": true, "env": true, "update": true, "completion": true}
	for _, c := range cmd.Commands() {
		delete(want, c.Name())
	}
	for name := range want {
		t.Errorf("missing subcommand %q", name)
	}
}

func TestCustomizeCompletionPT(t *testing.T) {
	d := testSelfDeps(t)
	cmd := NewSelfCmd(d)
	completion := findCommand(cmd.Commands(), "completion")
	if completion == nil {
		t.Fatal("completion command missing")
	}
	if completion.Short == "" {
		t.Error("expected Portuguese short on completion")
	}
	bash := findCommand(completion.Commands(), "bash")
	if bash == nil {
		t.Fatal("bash completion missing")
	}
	if bash.Short == "" || bash.GroupID == "" {
		t.Errorf("bash sub: Short=%q GroupID=%q", bash.Short, bash.GroupID)
	}
}
