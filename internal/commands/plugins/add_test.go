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
	if !strings.Contains(err.Error(), "nenhum plugin encontrado") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddCollectionTwoPlugins(t *testing.T) {
	d := testPluginsDeps(t)
	parent := t.TempDir()
	for _, pair := range []struct{ dir, cmd string }{{"alpha", "acmd"}, {"beta", "bcmd"}} {
		dir := filepath.Join(parent, pair.dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeMinimalRunnablePluginNamed(t, dir, pair.cmd)
	}
	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetArgs([]string{parent})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add collection: %v", err)
	}
	for _, name := range []string{"alpha", "beta"} {
		src, err := d.Store.GetPluginSource(name)
		if err != nil || src == nil {
			t.Fatalf("missing source %q: %v", name, err)
		}
	}
}

func TestAddCollectionNameWithMultipleFails(t *testing.T) {
	d := testPluginsDeps(t)
	parent := t.TempDir()
	for _, pair := range []struct{ dir, cmd string }{{"a", "acmd"}, {"b", "bcmd"}} {
		dir := filepath.Join(parent, pair.dir)
		_ = os.MkdirAll(dir, 0o755)
		writeMinimalRunnablePluginNamed(t, dir, pair.cmd)
	}
	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{parent, "--name", "x"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--name") {
		t.Fatalf("expected --name error: %v", err)
	}
}

func TestAddCollectionSkipsInvalidSubdir(t *testing.T) {
	d := testPluginsDeps(t)
	parent := t.TempDir()
	good := filepath.Join(parent, "good")
	_ = os.MkdirAll(good, 0o755)
	writeMinimalRunnablePlugin(t, good)
	bad := filepath.Join(parent, "bad")
	_ = os.MkdirAll(bad, 0o755)
	_ = os.WriteFile(filepath.Join(bad, "manifest.yaml"), []byte("command: x\ndescription: y\nentrypoint: missing.sh\n"), 0o644)

	var errBuf bytes.Buffer
	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{parent})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(errBuf.String(), "bad") {
		t.Errorf("expected stderr warn about bad: %s", errBuf.String())
	}
	src, _ := d.Store.GetPluginSource("good")
	if src == nil {
		t.Fatal("good should be registered")
	}
}

func TestAddCollectionSingleWithCustomName(t *testing.T) {
	d := testPluginsDeps(t)
	parent := t.TempDir()
	dir := filepath.Join(parent, "orig")
	_ = os.MkdirAll(dir, 0o755)
	writeMinimalRunnablePlugin(t, dir)
	cmd := newPluginsAddCmd(d)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(os.NewFile(0, os.DevNull))
	cmd.SetArgs([]string{parent, "--name", "custom"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Store.GetPluginSource("custom"); err != nil {
		t.Fatal(err)
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
