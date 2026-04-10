package envs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectVaultRowsDefaultOnly(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "__missing_mbcli.yaml"))
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectVaultRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Vault != "default" || rows[0].Path != def ||
		rows[0].EnvCount != 0 {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestCollectVaultRowsWithPerVaultFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "__missing_mbcli.yaml"))
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".env.staging"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectVaultRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len=%d", len(rows))
	}
	if rows[0].Vault != "default" || rows[0].Path != def || rows[0].EnvCount != 0 {
		t.Errorf("first: %+v", rows[0])
	}
	if rows[1].Vault != "staging" || rows[1].EnvCount != 0 {
		t.Errorf("second: %+v", rows[1])
	}
}

func TestCollectVaultRows_ignoresReservedDotEnvProject(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "__missing_mbcli.yaml"))
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".env.project"), []byte("X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectVaultRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Vault != "default" {
		t.Fatalf("want only default row, got %+v", rows)
	}
	for _, r := range rows {
		if r.Vault == "project" && r.Path == filepath.Join(tmp, ".env.project") {
			t.Fatalf("must not list file .env.project as config vault: %+v", rows)
		}
	}
}

func TestCollectVaultRows_projectVaultsFromMbcli(t *testing.T) {
	tmp := t.TempDir()
	mbcli := filepath.Join(tmp, "mbcli.yaml")
	t.Setenv("MBCLI_YAML_PATH", mbcli)
	t.Setenv("MBCLI_PROJECT_ROOT", "")
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("A=1\nB=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		mbcli,
		[]byte("envs:\n  P: yaml\n  stg:\n    Q: inner\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectVaultRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("want default+project+project/stg, got len=%d %+v", len(rows), rows)
	}
	if rows[0].Vault != "default" || rows[0].EnvCount != 2 {
		t.Fatalf("row0 %+v", rows[0])
	}
	if rows[1].Vault != "project" || rows[1].EnvCount != 1 || rows[1].Path != mbcli {
		t.Fatalf("row1 %+v", rows[1])
	}
	if rows[2].Vault != "project/stg" || rows[2].EnvCount != 1 {
		t.Fatalf("row2 %+v", rows[2])
	}
}
