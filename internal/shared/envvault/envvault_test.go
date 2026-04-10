package envvault

import (
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	if err := Validate("staging"); err != nil {
		t.Errorf("staging: %v", err)
	}
	if err := Validate("../x"); err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestFilePath(t *testing.T) {
	t.Parallel()
	cfg := "/home/u/.config/mb"
	p, err := FilePath(cfg, "staging")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(cfg, ".env.staging")
	if p != want {
		t.Errorf("got %q want %q", p, want)
	}
	_, err = FilePath(cfg, "../x")
	if err == nil {
		t.Fatal("expected error")
	}
}
