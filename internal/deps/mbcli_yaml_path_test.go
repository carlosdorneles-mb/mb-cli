package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMbcliYAMLPath_MBCLIYAMLPath(t *testing.T) {
	tmp := t.TempDir()
	want := filepath.Join(tmp, "custom.yaml")
	t.Setenv("MBCLI_YAML_PATH", want)
	t.Setenv("MBCLI_PROJECT_ROOT", "")
	got, err := ResolveMbcliYAMLPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveMbcliYAMLPath_RelativeYAMLPath(t *testing.T) {
	tmp := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MBCLI_YAML_PATH", "rel/mbcli.yaml")
	t.Setenv("MBCLI_PROJECT_ROOT", "")
	got, err := ResolveMbcliYAMLPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tmp, "rel", "mbcli.yaml")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveMbcliYAMLPath_ProjectRoot(t *testing.T) {
	tmp := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MBCLI_YAML_PATH", "")
	t.Setenv("MBCLI_PROJECT_ROOT", "subproj")
	got, err := ResolveMbcliYAMLPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tmp, "subproj", "mbcli.yaml")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
