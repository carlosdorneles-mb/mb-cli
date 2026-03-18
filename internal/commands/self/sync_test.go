package self

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/cache"
)

func TestRunSyncEmptyPluginsDir(t *testing.T) {
	d := testSelfDeps(t)
	var out bytes.Buffer
	err := RunSync(d, func(msg string) { out.WriteString(msg) }, &out)
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	if !strings.Contains(out.String(), "0 plugin") {
		t.Errorf("expected sync count message, got %q", out.String())
	}
}

func TestSelfSyncCmd(t *testing.T) {
	d := testSelfDeps(t)
	cmd := newSelfSyncCmd(d)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !strings.Contains(stdout.String(), "sincronizados") {
		t.Errorf("stdout: %s", stdout.String())
	}
}

func TestRunSyncPluginPathCollision(t *testing.T) {
	d := testSelfDeps(t)
	p1 := filepath.Join(d.Runtime.PluginsDir, "pkg1")
	p2 := filepath.Join(d.Runtime.PluginsDir, "pkg2")
	if err := os.MkdirAll(p1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p2, 0o755); err != nil {
		t.Fatal(err)
	}
	writePluginWithCommand(t, p1, "samecmd")
	writePluginWithCommand(t, p2, "samecmd")

	err := RunSync(d, nil, nil)
	if err == nil {
		t.Fatal("expected collision error")
	}
	if !strings.Contains(err.Error(), "conflito") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSyncRegistersLocalPathPlugin(t *testing.T) {
	d := testSelfDeps(t)
	pluginDir := t.TempDir()
	writePluginWithCommand(t, pluginDir, "fromlocal")
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "myloc",
		LocalPath:  pluginDir,
	}); err != nil {
		t.Fatal(err)
	}
	if err := RunSync(d, nil, nil); err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	plugins, err := d.Store.ListPlugins()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range plugins {
		if p.CommandName == "fromlocal" || strings.Contains(p.CommandPath, "fromlocal") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected plugin from local path, got %#v", plugins)
	}
}
