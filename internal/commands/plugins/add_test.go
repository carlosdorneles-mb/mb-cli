package plugincmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/commands/self"
	"mb/internal/executor"
	"mb/internal/plugins"
)

func TestDirHasManifest(t *testing.T) {
	tmp := t.TempDir()
	if dirHasManifest(tmp) {
		t.Error("empty dir should not have manifest")
	}
	if err := os.WriteFile(filepath.Join(tmp, "manifest.yaml"), []byte("command: test\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if !dirHasManifest(tmp) {
		t.Error("dir with manifest.yaml should have manifest")
	}
}

func TestAddLocalRegistersPathOnly(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	cachePath := filepath.Join(tmp, "cache.db")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	sourceDir := filepath.Join(tmp, "my-plugin")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte("command: hello\ndescription: Hi\ntype: sh\nentrypoint: run.sh\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	runtime := &config.RuntimeConfig{PluginsDir: pluginsDir}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	addCmd := newPluginsAddCmd(deps)
	addCmd.SetArgs([]string{"--name", "mylocal", sourceDir})
	addCmd.SetOut(os.NewFile(0, os.DevNull))
	addCmd.SetErr(os.NewFile(0, os.DevNull))
	if err := addCmd.Execute(); err != nil {
		t.Fatalf("add local: %v", err)
	}

	// No dir created in PluginsDir for local
	if _, err := os.Stat(filepath.Join(pluginsDir, "mylocal")); err == nil {
		t.Error("local plugin should not create dir in PluginsDir")
	}

	src, err := store.GetPluginSource("mylocal")
	if err != nil {
		t.Fatalf("get plugin source: %v", err)
	}
	if src == nil {
		t.Fatal("plugin source should exist after add local")
	}
	if src.LocalPath != sourceDir {
		t.Errorf("LocalPath want %q, got %q", sourceDir, src.LocalPath)
	}
	if src.GitURL != "" {
		t.Errorf("local plugin should have empty GitURL, got %q", src.GitURL)
	}
}

func TestAddLocalWithDotUsesCwd(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	cachePath := filepath.Join(tmp, "cache.db")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	sourceDir := filepath.Join(tmp, "cwd-plugin")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte("command: x\ndescription: X\ntype: sh\nentrypoint: run.sh\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	runtime := &config.RuntimeConfig{PluginsDir: pluginsDir}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	addCmd := newPluginsAddCmd(deps)
	addCmd.SetArgs([]string{".", "--name", "fromdot"})
	addCmd.SetOut(os.NewFile(0, os.DevNull))
	addCmd.SetErr(os.NewFile(0, os.DevNull))
	addCmd.SetIn(os.Stdin)
	prevWd, _ := os.Getwd()
	defer os.Chdir(prevWd)
	if err := os.Chdir(sourceDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := addCmd.Execute(); err != nil {
		t.Fatalf("add . : %v", err)
	}

	src, err := store.GetPluginSource("fromdot")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if src == nil || src.LocalPath != sourceDir {
		t.Errorf("expected LocalPath %q, got %#v", sourceDir, src)
	}
}

func TestAddLocalInvalidDirErrors(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	cachePath := filepath.Join(tmp, "cache.db")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	emptyDir := t.TempDir()
	runtime := &config.RuntimeConfig{PluginsDir: pluginsDir}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	addCmd := newPluginsAddCmd(deps)
	addCmd.SetArgs([]string{emptyDir})
	addCmd.SetOut(os.NewFile(0, os.DevNull))
	addCmd.SetErr(os.NewFile(0, os.DevNull))
	err = addCmd.Execute()
	if err == nil {
		t.Fatal("expected error when dir has no manifest.yaml")
	}
	if !strings.Contains(err.Error(), "manifest") {
		t.Errorf("error should mention manifest, got: %v", err)
	}
}

func TestSyncIncludesLocalSources(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	cachePath := filepath.Join(tmp, "cache.db")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	sourceDir := filepath.Join(tmp, "local-src")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "manifest.yaml"), []byte("command: cmd\ndescription: C\ntype: sh\nentrypoint: run.sh\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	if err := store.UpsertPluginSource(cache.PluginSource{InstallDir: "local-src", LocalPath: sourceDir}); err != nil {
		t.Fatalf("upsert source: %v", err)
	}

	runtime := &config.RuntimeConfig{PluginsDir: pluginsDir}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	if err := self.RunSync(deps, nil, nil); err != nil {
		t.Fatalf("sync: %v", err)
	}

	pluginList, err := store.ListPlugins()
	if err != nil {
		t.Fatalf("list plugins: %v", err)
	}
	var found bool
	for _, p := range pluginList {
		if p.CommandPath == "local-src" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sync should include plugin from local source; got %d plugins: %v", len(pluginList), pluginList)
	}
}
