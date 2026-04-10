package plugins

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/infra/sqlite"
	"mb/internal/shared/system"
	appplugins "mb/internal/usecase/plugins"
)

func TestRunSyncEmptyPluginsDir(t *testing.T) {
	d := testPluginsDeps(t)
	t.Setenv("PATH", t.TempDir())
	var buf bytes.Buffer
	log := system.NewLogger(false, false, &buf)
	_, err := RunSync(context.Background(), nil, d, log, appplugins.SyncOptions{EmitSuccess: true})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Nenhum comando alterado") &&
		!strings.Contains(out, "cache atualizado") {
		t.Errorf("expected no-change sync message, got %q", out)
	}
}

func TestRunSyncPluginPathCollision(t *testing.T) {
	d := testPluginsDeps(t)
	p1 := filepath.Join(d.Runtime.PluginsDir, "pkg1")
	p2 := filepath.Join(d.Runtime.PluginsDir, "pkg2")
	if err := os.MkdirAll(p1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p2, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMinimalRunnablePluginNamed(t, p1, "samecmd")
	writeMinimalRunnablePluginNamed(t, p2, "samecmd")

	_, err := RunSync(
		context.Background(),
		nil,
		d,
		system.NewLogger(false, false, io.Discard),
		appplugins.SyncOptions{},
	)
	if err == nil {
		t.Fatal("expected collision error")
	}
	if !strings.Contains(err.Error(), "conflito") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSyncRegistersLocalPathPlugin(t *testing.T) {
	d := testPluginsDeps(t)
	pluginDir := t.TempDir()
	writeMinimalRunnablePluginNamed(t, pluginDir, "fromlocal")
	if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "myloc",
		LocalPath:  pluginDir,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := RunSync(context.Background(), nil, d, nil, appplugins.SyncOptions{}); err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	pluginsList, err := d.Store.ListPlugins()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range pluginsList {
		if p.CommandName == "fromlocal" || strings.Contains(p.CommandPath, "fromlocal") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected plugin from local path, got %#v", pluginsList)
	}
}

func TestRunSyncClearsUnknownNestedGroupID(t *testing.T) {
	d := testPluginsDeps(t)
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
	if _, err := RunSync(context.Background(), nil, d, nil, appplugins.SyncOptions{}); err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	pluginsList, err := d.Store.ListPlugins()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pluginsList {
		if p.CommandName == "leaf" && p.GroupID != "" {
			t.Errorf("unknown group_id should be cleared in cache, got GroupID=%q", p.GroupID)
		}
	}
}

func TestRunSyncLogsUpdatedWhenOnlyReferencedScriptChanges(t *testing.T) {
	d := testPluginsDeps(t)
	t.Setenv("PATH", t.TempDir())
	pkg := filepath.Join(d.Runtime.PluginsDir, "p1")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMinimalRunnablePluginNamed(t, pkg, "scriptonly")

	var buf1 bytes.Buffer
	log1 := system.NewLogger(false, false, &buf1)
	if _, err := RunSync(context.Background(), nil, d, log1, appplugins.SyncOptions{}); err != nil {
		t.Fatalf("RunSync 1: %v", err)
	}
	if !strings.Contains(buf1.String(), "adicionado") {
		t.Fatalf("expected added command log, got %q", buf1.String())
	}

	runSh := filepath.Join(pkg, "run.sh")
	if err := os.WriteFile(runSh, []byte("#!/bin/sh\necho changed\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	var buf2 bytes.Buffer
	log2 := system.NewLogger(false, false, &buf2)
	if _, err := RunSync(context.Background(), nil, d, log2, appplugins.SyncOptions{}); err != nil {
		t.Fatalf("RunSync 2: %v", err)
	}
	out := buf2.String()
	if !strings.Contains(out, "atualizado") || !strings.Contains(out, "scriptonly") {
		t.Fatalf("expected update log when only run.sh changed, got %q", out)
	}
}
