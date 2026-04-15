package aliases

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_missingIsEmpty(t *testing.T) {
	f, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Version != currentVersion || len(f.Aliases) != 0 {
		t.Fatalf("got %#v", f)
	}
}

func TestSave_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aliases.yaml")
	f := &File{
		Version: 1,
		Aliases: map[string]Entry{
			"x": {Command: []string{"docker", "ps"}, EnvVault: "st"},
		},
	}
	if err := Save(path, f); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Aliases) != 1 || got.Aliases["x"].EnvVault != "st" {
		t.Fatalf("got %#v", got.Aliases)
	}
	if len(got.Aliases["x"].Command) != 2 {
		t.Fatalf("command: %#v", got.Aliases["x"].Command)
	}
}

func TestValidateName(t *testing.T) {
	if err := ValidateName("ok-1"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateName("bad name"); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveForRun(t *testing.T) {
	dir := t.TempDir()
	path := FilePath(dir)
	raw := []byte(`version: 1
aliases:
  doit:
    command: ["true"]
  vonly:
    command: ["echo", "x"]
    env_vault: staging
`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	argv, v, ok, err := ResolveForRun(dir, "doit", nil)
	if err != nil || !ok || len(argv) != 1 || argv[0] != "true" || v != "" {
		t.Fatalf("got argv=%v v=%q ok=%v err=%v", argv, v, ok, err)
	}
	argv, v, ok, err = ResolveForRun(dir, "missing", []string{"a"})
	if err != nil || ok || argv[0] != "missing" || argv[1] != "a" || v != "" {
		t.Fatalf("got argv=%v v=%q ok=%v err=%v", argv, v, ok, err)
	}
	argv, v, ok, err = ResolveForRun(dir, "vonly", []string{"extra"})
	if err != nil || !ok || v != "staging" || len(argv) != 3 || argv[2] != "extra" {
		t.Fatalf("got argv=%v v=%q ok=%v err=%v", argv, v, ok, err)
	}
}

func TestEffectiveEnvVault(t *testing.T) {
	if got := EffectiveEnvVault("cli", "alias"); got != "cli" {
		t.Fatalf("got %q", got)
	}
	if got := EffectiveEnvVault("", "alias"); got != "alias" {
		t.Fatalf("got %q", got)
	}
}
