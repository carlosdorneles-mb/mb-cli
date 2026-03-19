package self

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/cache"
	"mb/internal/system"
)

func TestRunSyncEmptyPluginsDir(t *testing.T) {
	d := testSelfDeps(t)
	t.Setenv("PATH", t.TempDir())
	var buf bytes.Buffer
	log := system.NewLogger(false, false, &buf)
	err := RunSync(context.Background(), d, log, true)
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	if !strings.Contains(buf.String(), "0 plugin") &&
		!strings.Contains(buf.String(), "sincronizados") {
		t.Errorf("expected sync count message, got %q", buf.String())
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

	err := RunSync(context.Background(), d, system.NewLogger(false, false, io.Discard), false)
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
	if err := RunSync(context.Background(), d, nil, false); err != nil {
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

func TestRunSyncClearsUnknownNestedGroupID(t *testing.T) {
	d := testSelfDeps(t)
	pkg := filepath.Join(d.Runtime.PluginsDir, "nest")
	if err := os.MkdirAll(filepath.Join(pkg, "sub", "leaf"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(pkg, "manifest.yaml"),
		[]byte("command: pkg\ndescription: r\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(pkg, "sub", "manifest.yaml"),
		[]byte("command: sub\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(pkg, "sub", "leaf", "manifest.yaml"),
		[]byte("command: leaf\ndescription: l\ngroup_id: not_registered\nentrypoint: run.sh\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(pkg, "sub", "leaf", "run.sh"),
		[]byte("#!/bin/sh\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := RunSync(context.Background(), d, nil, false); err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	plugins, err := d.Store.ListPlugins()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range plugins {
		if p.CommandName == "leaf" && p.GroupID != "" {
			t.Errorf("unknown group_id should be cleared in cache, got GroupID=%q", p.GroupID)
		}
	}
}
