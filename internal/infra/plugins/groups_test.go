package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGroupsFile_valid(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "groups.yaml")
	if err := os.WriteFile(p, []byte(`
- id: k8s_ops
  title: KUBERNETES
- id: ci_cd
  title: CI/CD
`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadGroupsFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "k8s_ops" || got[1].ID != "ci_cd" {
		t.Fatalf("got %+v", got)
	}
}

func TestLoadGroupsFile_missing(t *testing.T) {
	got, err := LoadGroupsFile(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil || got != nil {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestLoadGroupsFile_duplicateID(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "groups.yaml")
	_ = os.WriteFile(p, []byte(`
- id: a
  title: A
- id: a
  title: B
`), 0o644)
	_, err := LoadGroupsFile(p)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadGroupsFile_reservedID(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "groups.yaml")
	_ = os.WriteFile(p, []byte(`
- id: commands
  title: X
`), 0o644)
	_, err := LoadGroupsFile(p)
	if err == nil {
		t.Fatal("expected error")
	}
}
