package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

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

func TestParseMbcliAliases_shorthandLists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := `aliases:
  pinga:
    - echo
    - hi
  staging:
    test:
      - mb
      - plugins
      - test
`
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m[alib.StoreKey("", "pinga")]; got.EnvVault != "" || len(got.Command) != 2 ||
		got.Command[0] != "echo" {
		t.Fatalf("pinga=%+v", got)
	}
	if got := m[alib.StoreKey("staging", "test")]; got.EnvVault != "staging" ||
		got.Command[0] != "mb" ||
		got.Command[2] != "test" {
		t.Fatalf("test=%+v", got)
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
	if got := m[alib.StoreKey("", "root1")]; got.EnvVault != "" || len(got.Command) != 2 ||
		got.Command[1] != "root" {
		t.Fatalf("root1=%+v", got)
	}
	if got := m[alib.StoreKey("staging", "inner1")]; got.EnvVault != "staging" ||
		got.Command[1] != "inner" {
		t.Fatalf("inner1=%+v", got)
	}
	if got := m[alib.StoreKey("prod", "inner2")]; got.EnvVault != "prod" {
		t.Fatalf("inner2=%+v", got)
	}
}

func TestParseMbcliAliases_sameDisplayNameTwoVaults_ok(t *testing.T) {
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
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m[alib.StoreKey("", "dup")]; len(got.Command) != 1 || got.Command[0] != "a" {
		t.Fatalf("root dup=%+v", got)
	}
	if got := m[alib.StoreKey("staging", "dup")]; len(got.Command) != 1 || got.Command[0] != "b" {
		t.Fatalf("staging dup=%+v", got)
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
	if got := m[alib.StoreKey("", "inner")]; got.EnvVault != "" {
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
	if got := m[alib.StoreKey("st", "t1")]; got.EnvVault != "st" || len(got.Command) != 1 {
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
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	aliases, _ := root["aliases"].(map[string]any)
	if aliases == nil {
		t.Fatal("expected aliases map")
	}
	st, ok := aliases["st"].(map[string]any)
	if !ok {
		t.Fatalf("expected vault st as nested map, got %T %v", aliases["st"], aliases["st"])
	}
	t1v, ok := st["t1"].([]any)
	if !ok || len(t1v) != 1 {
		t.Fatalf("expected st.t1 as argv list, got %#v", st["t1"])
	}
	if err := RemoveMbcliYAMLAlias(p, "t1", "st"); err != nil {
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

func TestUpsertMbcliYAML_shorthandMixedRoundTrip(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "mbcli.yaml")
	if err := os.WriteFile(p, []byte("envs:\n  X: \"y\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := UpsertMbcliYAMLAlias(
		p,
		"rootA",
		alib.Entry{Command: []string{"echo", "r"}},
	); err != nil {
		t.Fatal(err)
	}
	if err := UpsertMbcliYAMLAlias(
		p,
		"inSt",
		alib.Entry{Command: []string{"true"}, EnvVault: "staging"},
	); err != nil {
		t.Fatal(err)
	}
	m, err := ParseMbcliAliases(p)
	if err != nil {
		t.Fatal(err)
	}
	if m[alib.StoreKey("", "rootA")].EnvVault != "" ||
		m[alib.StoreKey("staging", "inSt")].EnvVault != "staging" {
		t.Fatalf("%+v %+v", m[alib.StoreKey("", "rootA")], m[alib.StoreKey("staging", "inSt")])
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	aliases := root["aliases"].(map[string]any)
	if _, ok := aliases["rootA"].([]any); !ok {
		t.Fatalf("want rootA as list, got %T", aliases["rootA"])
	}
	st := aliases["staging"].(map[string]any)
	if _, ok := st["inSt"].([]any); !ok {
		t.Fatalf("want staging.inSt as list, got %T", st["inSt"])
	}
	if !strings.Contains(string(data), "envs") {
		t.Fatal("lost envs")
	}
}

func TestUpsertMbcliYAMLAlias_rootNameVaultClash(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "mbcli.yaml")
	if err := UpsertMbcliYAMLAlias(
		p,
		"inner",
		alib.Entry{Command: []string{"id"}, EnvVault: "staging"},
	); err != nil {
		t.Fatal(err)
	}
	err := UpsertMbcliYAMLAlias(p, "staging", alib.Entry{Command: []string{"echo", "x"}})
	if err == nil {
		t.Fatal("expected clash error")
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
	if len(m) != 1 || m[alib.StoreKey("", "x")].Command[0] != "id" {
		t.Fatalf("%v", m)
	}
}
