package plugins

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/cache"
)

func TestAddRequiresExactlyOneArg(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error with no args")
	}

	cmd.SetArgs([]string{"a", "b"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error with two args")
	}
}

func TestAddLocalPathNotFound(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := newPluginsAddCmd(d)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	missing := filepath.Join(t.TempDir(), "nope")
	cmd.SetArgs([]string{missing})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "não encontrado") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddLocalPathNotDirectory(t *testing.T) {
	d := testPluginsDeps(t)
	cmd := newPluginsAddCmd(d)
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "não é um diretório") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddLocalNoManifest(t *testing.T) {
	d := testPluginsDeps(t)
	emptyDir := t.TempDir()
	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{emptyDir})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "manifest.yaml") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddLocalRegistersPlugin(t *testing.T) {
	d := testPluginsDeps(t)
	pluginDir := t.TempDir()
	writeMinimalRunnablePlugin(t, pluginDir)

	cmd := newPluginsAddCmd(d)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetArgs([]string{pluginDir, "--name", "myplug"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	src, err := d.Store.GetPluginSource("myplug")
	if err != nil {
		t.Fatalf("GetPluginSource: %v", err)
	}
	if src == nil {
		t.Fatal("expected plugin source")
	}
	if src.LocalPath == "" {
		t.Error("expected LocalPath set")
	}
	if !strings.Contains(out.String(), "myplug") {
		t.Errorf("stdout should mention plugin name: %s", out.String())
	}
}

func TestAddLocalDuplicateName(t *testing.T) {
	d := testPluginsDeps(t)
	if err := d.Store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "taken",
		LocalPath:  "/tmp/x",
	}); err != nil {
		t.Fatal(err)
	}

	pluginDir := t.TempDir()
	writeMinimalRunnablePlugin(t, pluginDir)

	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{pluginDir, "--name", "taken"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "já existe") {
		t.Errorf("unexpected error: %v", err)
	}
}
