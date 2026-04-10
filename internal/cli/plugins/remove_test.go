package plugins

import (
	"bytes"
	"strings"
	"testing"

	"mb/internal/infra/sqlite"
)

func forceConfirmFallback(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
}

func TestRemovePluginNotFound(t *testing.T) {
	_, _, rmSvc, _, d := testAllPluginServicesWithDeps(t)
	cmd := newPluginsRemoveCmd(rmSvc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetIn(strings.NewReader("y\n"))
	cmd.SetArgs([]string{"ghost"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "não encontrado") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRemoveCancelled(t *testing.T) {
	forceConfirmFallback(t)
	_, _, rmSvc, _, d := testAllPluginServicesWithDeps(t)
	if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "keepme",
		LocalPath:  "/some/path",
	}); err != nil {
		t.Fatal(err)
	}

	var errBuf bytes.Buffer
	cmd := newPluginsRemoveCmd(rmSvc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errBuf)
	cmd.SetIn(strings.NewReader("n\n"))
	cmd.SetArgs([]string{"keepme"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(errBuf.String(), "cancelada") {
		t.Errorf("expected cancel message, got: %s", errBuf.String())
	}
	src, err := d.Store.GetPluginSource("keepme")
	if err != nil {
		t.Fatal(err)
	}
	if src == nil {
		t.Fatal("plugin should still be registered after cancel")
	}
}

func TestRemoveLocalConfirmed(t *testing.T) {
	forceConfirmFallback(t)
	_, _, rmSvc, _, d := testAllPluginServicesWithDeps(t)
	if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "gone",
		LocalPath:  "/tmp/plugin",
	}); err != nil {
		t.Fatal(err)
	}

	var errBuf bytes.Buffer
	cmd := newPluginsRemoveCmd(rmSvc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errBuf)
	cmd.SetIn(strings.NewReader("yes\n"))
	cmd.SetArgs([]string{"gone"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	src, err := d.Store.GetPluginSource("gone")
	if err != nil {
		t.Fatal(err)
	}
	if src != nil {
		t.Error("plugin should be gone from registry")
	}
}
