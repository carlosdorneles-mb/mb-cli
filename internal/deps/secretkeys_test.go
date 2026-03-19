package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSecretKeys_NotFound(t *testing.T) {
	tmp := t.TempDir()
	keys, err := LoadSecretKeys(filepath.Join(tmp, "env.defaults"))
	if err != nil {
		t.Fatal(err)
	}
	if keys != nil {
		t.Errorf("expected nil, got %v", keys)
	}
}

func TestLoadSecretKeys_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p+".secrets", []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSecretKeys(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty, got %v", keys)
	}
}

func TestLoadSecretKeys_WithKeys(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p+".secrets", []byte("API_KEY\nDB_PASS\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSecretKeys(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "API_KEY" || keys[1] != "DB_PASS" {
		t.Errorf("got %v", keys)
	}
}

func TestAddSecretKey_NewFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := AddSecretKey(p, "TOKEN"); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSecretKeys(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "TOKEN" {
		t.Errorf("got %v", keys)
	}
}

func TestAddSecretKey_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := AddSecretKey(p, "TOKEN"); err != nil {
		t.Fatal(err)
	}
	if err := AddSecretKey(p, "TOKEN"); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSecretKeys(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "TOKEN" {
		t.Errorf("got %v", keys)
	}
}

func TestRemoveSecretKey(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := AddSecretKey(p, "A"); err != nil {
		t.Fatal(err)
	}
	if err := AddSecretKey(p, "B"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveSecretKey(p, "A"); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSecretKeys(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "B" {
		t.Errorf("got %v", keys)
	}
}

func TestRemoveSecretKey_AllRemoved_DeletesFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := AddSecretKey(p, "X"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveSecretKey(p, "X"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p + ".secrets"); !os.IsNotExist(err) {
		t.Errorf("expected .secrets file to be removed")
	}
}
