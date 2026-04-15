package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertMbcliYAMLEnvs_root_preservesAliases(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	if err := os.WriteFile(
		p,
		[]byte("aliases:\n  x:\n    command: [echo]\nenvs:\n  A: \"1\"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := UpsertMbcliYAMLEnvs(p, "", map[string]string{"B": "2"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "aliases") || !strings.Contains(s, "A:") || !strings.Contains(s, "B:") {
		t.Fatalf("expected aliases+A+B: %s", s)
	}
}

func TestUpsertMbcliYAMLEnvs_nested(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	if err := UpsertMbcliYAMLEnvs(p, "staging", map[string]string{"X": "y"}); err != nil {
		t.Fatal(err)
	}
	def, byV, err := ParseMbcliProjectEnvs(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(def) != 0 {
		t.Fatalf("def=%v want empty", def)
	}
	if byV["staging"]["X"] != "y" {
		t.Fatalf("byV staging X=%q", byV["staging"]["X"])
	}
}

func TestUpsertMbcliYAMLEnvs_conflictRootKeyIsNestedVault(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  staging:\n    FOO: bar\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	err := UpsertMbcliYAMLEnvs(p, "", map[string]string{"staging": "scalar"})
	if err == nil || !strings.Contains(err.Error(), "vault aninhado") {
		t.Fatalf("want nested vault error, got %v", err)
	}
}

func TestUpsertMbcliYAMLEnvs_conflictVaultIsScalar(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  staging: plain\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	err := UpsertMbcliYAMLEnvs(p, "staging", map[string]string{"K": "v"})
	if err == nil || !strings.Contains(err.Error(), "valor escalar") {
		t.Fatalf("want scalar conflict error, got %v", err)
	}
}

func TestRemoveMbcliYAMLEnvKeys_nestedEmptiesVault(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  staging:\n    ONLY: \"1\"\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RemoveMbcliYAMLEnvKeys(p, "staging", []string{"ONLY"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "staging:") {
		t.Fatalf("expected staging vault removed: %s", string(b))
	}
}

func TestUpsertMbcliYAMLEnvs_rejectsVaultProject(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	err := UpsertMbcliYAMLEnvs(p, "project", map[string]string{"K": "v"})
	if err == nil || !strings.Contains(err.Error(), "reservados") {
		t.Fatalf("want reserved vault error, got %v", err)
	}
}

func TestRemoveMbcliYAMLEnvKeys_missing(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	if err := os.WriteFile(p, []byte("envs:\n  A: \"1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := RemoveMbcliYAMLEnvKeys(p, "", []string{"A", "NOPE"})
	if err == nil || !strings.Contains(err.Error(), "inexistentes") {
		t.Fatalf("want missing error, got %v", err)
	}
	def, _, _ := ParseMbcliProjectEnvs(p)
	if def["A"] != "1" {
		t.Fatalf("A should remain: %v", def)
	}
}
