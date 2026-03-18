package plugins

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"mb/internal/cache"
)

func TestRemovePluginNotFound(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := newPluginsRemoveCmd(d)
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
	d := testPluginsDeps(t)
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "keepme",
		LocalPath:  "/some/path",
	}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := newPluginsRemoveCmd(d)
	cmd.SetOut(&out)
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetIn(strings.NewReader("n\n"))
	cmd.SetArgs([]string{"keepme"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "cancelada") {
		t.Errorf("expected cancel message, got: %s", out.String())
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
	d := testPluginsDeps(t)
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "gone",
		LocalPath:  "/tmp/plugin",
	}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := newPluginsRemoveCmd(d)
	cmd.SetOut(&out)
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetIn(strings.NewReader("yes\n"))
	cmd.SetArgs([]string{"gone"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "removido") {
		t.Errorf("expected success message: %s", out.String())
	}
	src, err := d.Store.GetPluginSource("gone")
	if err != nil {
		t.Fatal(err)
	}
	if src != nil {
		t.Fatal("plugin source should be deleted")
	}
}

func TestConfirmRemoveYesShort(t *testing.T) {
	cmd := newPluginsRemoveCmd(testPluginsDeps(t))
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetIn(strings.NewReader("y\n"))
	ok, err := confirmRemove(cmd, "p")
	if err != nil || !ok {
		t.Fatalf("confirmRemove y: ok=%v err=%v", ok, err)
	}
}

func TestConfirmRemoveEmptyLine(t *testing.T) {
	cmd := newPluginsRemoveCmd(testPluginsDeps(t))
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetIn(strings.NewReader("\n"))
	ok, err := confirmRemove(cmd, "p")
	if err != nil || ok {
		t.Fatalf("empty line should be N: ok=%v err=%v", ok, err)
	}
}
