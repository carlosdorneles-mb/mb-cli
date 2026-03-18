package plugins

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"mb/internal/cache"
)

func TestUpdateRequiresNameOrAll(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := newPluginsUpdateCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error without args and without --all")
	}
	if !strings.Contains(err.Error(), "--all") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdatePluginNotFound(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := newPluginsUpdateCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"missing"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "não encontrado") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateLocalPlugin(t *testing.T) {
	d := testPluginsDeps(t)
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "loc",
		LocalPath:  "/tmp/x",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newPluginsUpdateCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"loc"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "local") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateManualInstallNoGitURL(t *testing.T) {
	d := testPluginsDeps(t)
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "manual",
		GitURL:     "",
		LocalPath:  "",
		RefType:    "tag",
		Ref:        "v1",
		Version:    "v1",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newPluginsUpdateCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"manual"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "manualmente") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateAllWithOnlyLocalPlugins(t *testing.T) {
	d := testPluginsDeps(t)
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "loc",
		LocalPath:  "/tmp/p",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newPluginsUpdateCmd(d)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update --all: %v", err)
	}
}
