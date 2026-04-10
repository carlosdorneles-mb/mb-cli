package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMbcliProjectEnvs_missingFile(t *testing.T) {
	def, byV, err := ParseMbcliProjectEnvs(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(def) != 0 || len(byV) != 0 {
		t.Fatalf("want empty, got def=%#v byV=%#v", def, byV)
	}
}

func TestParseMbcliProjectEnvs_nonMapEnvs(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	if err := os.WriteFile(p, []byte("envs: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := ParseMbcliProjectEnvs(p)
	if err == nil {
		t.Fatal("expected error for sequence envs")
	}
}

func TestParseMbcliProjectEnvs_invalidVaultName(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  'bad name':\n    X: '1'\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := ParseMbcliProjectEnvs(p)
	if err == nil {
		t.Fatal("expected error for invalid nested vault name")
	}
}

func TestParseMbcliProjectEnvs_nestedVault(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  FOO: root\n  staging:\n    FOO: stg\n    BAR: 2\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	def, byV, err := ParseMbcliProjectEnvs(p)
	if err != nil {
		t.Fatal(err)
	}
	if def["FOO"] != "root" {
		t.Fatalf("def FOO=%q", def["FOO"])
	}
	if byV["staging"]["FOO"] != "stg" || byV["staging"]["BAR"] != "2" {
		t.Fatalf("staging=%#v", byV["staging"])
	}
}

func TestLoadMbcliProjectEnvsForMerge_noVault(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  FOO: root\n  staging:\n    FOO: stg\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadMbcliProjectEnvsForMerge(p, "")
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "root" {
		t.Fatalf("FOO=%q want root (staging overlay sem --env-vault)", m["FOO"])
	}
}

func TestLoadMbcliProjectEnvsForMerge_withVault(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  FOO: root\n  staging:\n    FOO: stg\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadMbcliProjectEnvsForMerge(p, "staging")
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "stg" {
		t.Fatalf("FOO=%q want stg", m["FOO"])
	}
}

func TestMbcliProjectListEntries_filterDefault(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  A: '1'\n  staging:\n    B: '2'\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := MbcliProjectListEntries(p, "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Key != "A" || rows[0].Vault != "project" {
		t.Fatalf("%#v", rows)
	}
}

func TestMbcliProjectListEntries_projectSlashStagingNestedOnly(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  A: root\n  staging:\n    B: inner\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := MbcliProjectListEntries(p, "project/staging")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Key != "B" || rows[0].Vault != "project/staging" {
		t.Fatalf("%#v", rows)
	}
}

func TestMbcliProjectListEntries_filterStaging(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  A: '1'\n  staging:\n    B: '2'\n"
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := MbcliProjectListEntries(p, "staging")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len=%d %#v", len(rows), rows)
	}
}
