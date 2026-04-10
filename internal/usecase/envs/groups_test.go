package envs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectEnvGroupRowsDefaultOnly(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectEnvGroupRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Group != "default" || rows[0].Path != def {
		t.Fatalf("got %+v", rows)
	}
}

func TestCollectEnvGroupRowsWithPerGroupFiles(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(filepath.Join(tmp, ".env.staging"), []byte("X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(tmp, ".env.prod.secrets"),
		[]byte("K\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	paths := Paths{DefaultEnvPath: def, ConfigDir: tmp}
	rows, err := CollectEnvGroupRows(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len=%d rows=%+v", len(rows), rows)
	}
	if rows[0].Group != "default" || rows[0].Path != def {
		t.Fatalf("first row: %+v", rows[0])
	}
	if rows[1].Group != "staging" {
		t.Fatalf("second row: %+v", rows[1])
	}
}
