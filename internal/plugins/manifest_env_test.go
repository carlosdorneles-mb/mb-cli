package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeManifestEnvFiles_EmptyJSON(t *testing.T) {
	m, err := MergeManifestEnvFiles("/tmp", "", ManifestEnvGroupDefault)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Fatalf("want empty map, got %v", m)
	}
}

func TestMergeManifestEnvFiles_GroupFilterAndOrder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, ".env.default"),
		[]byte("A=1\nB=1\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.test.a"), []byte("B=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.test.b"), []byte("C=3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	json := `[{"file":".env.default","group":"default"},{"file":".env.test.a","group":"test"},{"file":".env.test.b","group":"test"}]`

	m, err := MergeManifestEnvFiles(dir, json, "test")
	if err != nil {
		t.Fatal(err)
	}
	if m["B"] != "2" || m["C"] != "3" || m["A"] != "" {
		t.Fatalf("test group: got %+v want B=2 C=3 no A", m)
	}

	m2, err := MergeManifestEnvFiles(dir, json, ManifestEnvGroupDefault)
	if err != nil {
		t.Fatal(err)
	}
	if m2["A"] != "1" || m2["B"] != "1" || m2["C"] != "" {
		t.Fatalf("default group: got %+v", m2)
	}
}

func TestMergeManifestEnvFiles_MissingFile(t *testing.T) {
	dir := t.TempDir()
	json := `[{"file":"missing.env","group":"default"}]`
	_, err := MergeManifestEnvFiles(dir, json, ManifestEnvGroupDefault)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestMergeManifestEnvFiles_LastFileWins(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.env"), []byte("X=1\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "b.env"), []byte("X=2\n"), 0o600)
	json := `[{"file":"a.env","group":"staging"},{"file":"b.env","group":"staging"}]`
	m, err := MergeManifestEnvFiles(dir, json, "staging")
	if err != nil {
		t.Fatal(err)
	}
	if m["X"] != "2" {
		t.Errorf("X=%q want 2", m["X"])
	}
}
