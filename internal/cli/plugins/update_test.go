package plugins

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"mb/internal/infra/sqlite"
)

func TestUpdateRequiresPackageOrAll(t *testing.T) {
	_, _, _, upSvc, d := testAllPluginServicesWithDeps(t)
	cmd := newPluginsUpdateCmd(upSvc, d)
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
	_, _, _, upSvc, d := testAllPluginServicesWithDeps(t)
	cmd := newPluginsUpdateCmd(upSvc, d)
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
	_, _, _, upSvc, d := testAllPluginServicesWithDeps(t)
	if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "loc",
		LocalPath:  "/tmp/x",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newPluginsUpdateCmd(upSvc, d)
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

func TestUpdateRemoteMissingGitURL(t *testing.T) {
	_, _, _, upSvc, d := testAllPluginServicesWithDeps(t)
	if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "nogit",
		LocalPath:  "",
		GitURL:     "",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newPluginsUpdateCmd(upSvc, d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"nogit"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "manualmente") && !strings.Contains(err.Error(), "URL") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateAllWithLocalPlugins(t *testing.T) {
	_, _, _, upSvc, d := testAllPluginServicesWithDeps(t)
	for _, name := range []string{"l1", "l2"} {
		if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
			InstallDir: name,
			LocalPath:  "/tmp/" + name,
		}); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newPluginsUpdateCmd(upSvc, d)
	var errBuf bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Logf("stderr: %s", errBuf.String())
		t.Fatalf("execute --all: %v", err)
	}
}

func TestUpdateNonGitPluginSkipped(t *testing.T) {
	_, _, _, upSvc, d := testAllPluginServicesWithDeps(t)
	if err := d.Store.UpsertPluginSource(sqlite.PluginSource{
		InstallDir: "skipped",
		LocalPath:  "/opt/plugin",
		GitURL:     "",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newPluginsUpdateCmd(upSvc, d)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Logf("update --all with non-git: %v", err)
	}
}
