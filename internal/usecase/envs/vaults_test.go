package envs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectVaultRowsDefaultOnly(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectVaultRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Vault != "default" || rows[0].Path != def {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestCollectVaultRowsWithPerVaultFiles(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
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
	if rows[0].Vault != "default" || rows[0].Path != def {
		t.Errorf("first: %+v", rows[0])
	}
	if rows[1].Vault != "staging" {
		t.Errorf("second: %+v", rows[1])
	}
}
