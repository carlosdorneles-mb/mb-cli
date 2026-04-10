package deps

import (
	"path/filepath"
	"testing"
)

func TestValidateEnvVault(t *testing.T) {
	t.Parallel()
	if err := ValidateEnvVault("staging"); err != nil {
		t.Errorf("staging: %v", err)
	}
	if err := ValidateEnvVault("../x"); err == nil {
		t.Error("expected error")
	}
}

func TestVaultEnvFilePath(t *testing.T) {
	t.Parallel()
	cfg := t.TempDir()
	p, err := VaultEnvFilePath(cfg, "staging")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(cfg, ".env.staging")
	if p != want {
		t.Errorf("got %q want %q", p, want)
	}
	_, err = VaultEnvFilePath(cfg, "../x")
	if err == nil {
		t.Fatal("expected error")
	}
}
