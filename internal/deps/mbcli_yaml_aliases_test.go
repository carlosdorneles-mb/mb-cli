package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	alib "mb/internal/shared/aliases"
)

func TestParseMbcliAliases_missingFile(t *testing.T) {
	t.Parallel()
	m, err := ParseMbcliAliases(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Fatalf("got %v", m)
	}
}

func TestParseMbcliAliases_flatAndNested(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := `aliases:
  root1:
    command: [echo, root]
  staging:
    inner1:
      command: [echo, inner]
    inner2:
      command: [echo, x]
      env_vault: prod
`
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m["root1"]; got.EnvVault != "" || len(got.Command) != 2 || got.Command[1] != "root" {
		t.Fatalf("root1=%+v", got)
	}
	if got := m["inner1"]; got.EnvVault != "staging" || got.Command[1] != "inner" {
		t.Fatalf("inner1=%+v", got)
	}
	if got := m["inner2"]; got.EnvVault != "prod" {
		t.Fatalf("inner2=%+v", got)
	}
}

func TestParseMbcliAliases_duplicateFlattened(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := `aliases:
  dup:
    command: [a]
  staging:
    dup:
      command: [b]
`
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseMbcliAliases(p)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestParseMbcliAliases_reservedProjectGroup(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := `aliases:
  project:
    x:
      command: [echo]
`
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseMbcliAliases(p)
	if err == nil {
		t.Fatal("expected error for reserved group project")
	}
}

func TestParseMbcliAliases_reservedTopLevelProject(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := `aliases:
  project:
    command: [echo]
`
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseMbcliAliases(p)
	if err == nil {
		t.Fatal("expected error for reserved top-level key project")
	}
}

func TestParseMbcliAliases_explicitEmptyEnvVaultClearsImplicit(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := `aliases:
  staging:
    inner:
      command: [echo, x]
      env_vault: ""
`
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m["inner"]; got.EnvVault != "" {
		t.Fatalf("want empty env_vault, got %+v", got)
	}
}

func TestUpsertRemoveMbcliYAMLRoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	if err := os.WriteFile(p, []byte("envs:\n  A: \"1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := alib.Entry{Command: []string{"true"}, EnvVault: "st"}
	if err := UpsertMbcliYAMLAlias(p, "t1", e); err != nil {
		t.Fatal(err)
	}
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m["t1"]; got.EnvVault != "st" || len(got.Command) != 1 {
		t.Fatalf("got %+v", got)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "envs") || !strings.Contains(s, "aliases") {
		t.Fatalf("expected envs+aliases preserved: %s", s)
	}
	if err := RemoveMbcliYAMLAlias(p, "t1"); err != nil {
		t.Fatal(err)
	}
	m2, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(m2) != 0 {
		t.Fatalf("aliases left: %v", m2)
	}
}

func TestUpsertMbcliYAMLAlias_createsFile(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "mbcli.yaml")
	if err := UpsertMbcliYAMLAlias(p, "x", alib.Entry{Command: []string{"id"}}); err != nil {
		t.Fatal(err)
	}
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 || m["x"].Command[0] != "id" {
		t.Fatalf("%v", m)
	}
}
