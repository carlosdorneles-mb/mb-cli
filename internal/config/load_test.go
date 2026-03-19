package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"mb/internal/version"
)

func TestLoad_CreatesConfigWhenMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("config should not exist yet")
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config.yaml should be created: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("docs_url:")) || bytes.Contains(raw, []byte("update_repo:")) {
		t.Fatalf("created config should not embed default key values:\n%s", raw)
	}
	if cfg.DocsBaseURL != DefaultDocsURL {
		t.Errorf("DocsBaseURL = %q", cfg.DocsBaseURL)
	}
	want := DefaultUpdateRepo
	if version.UpdateRepo != "" {
		want = version.UpdateRepo
	}
	if cfg.UpdateRepo != want {
		t.Errorf("UpdateRepo = %q, want %q", cfg.UpdateRepo, want)
	}
}

func TestLoad_WithUpdateRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "config.yaml"),
		[]byte("docs_url: https://example.com/\nupdate_repo: fork/mb-cli\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UpdateRepo != "fork/mb-cli" {
		t.Errorf("UpdateRepo = %q", cfg.UpdateRepo)
	}
	if cfg.DocsBaseURL != "https://example.com/" {
		t.Errorf("DocsBaseURL = %q", cfg.DocsBaseURL)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "config.yaml"),
		[]byte("update_repo: [\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_InvalidUpdateRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "config.yaml"),
		[]byte("docs_url: https://a.com/\nupdate_repo: noshslashes\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_InvalidDocsURL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "config.yaml"),
		[]byte("docs_url: not-a-url\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_OtherKeysOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "config.yaml"),
		[]byte("other: true\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DocsBaseURL != DefaultDocsURL {
		t.Errorf("DocsBaseURL = %q, want default", cfg.DocsBaseURL)
	}
	want := DefaultUpdateRepo
	if version.UpdateRepo != "" {
		want = version.UpdateRepo
	}
	if cfg.UpdateRepo != want {
		t.Errorf("UpdateRepo = %q, want %q", cfg.UpdateRepo, want)
	}
}

func TestValidateUpdateRepo(t *testing.T) {
	t.Parallel()
	if err := ValidateUpdateRepo("org/repo"); err != nil {
		t.Error(err)
	}
	if err := ValidateUpdateRepo(""); err == nil {
		t.Error("want error")
	}
	if err := ValidateUpdateRepo("bad"); err == nil {
		t.Error("want error")
	}
}
