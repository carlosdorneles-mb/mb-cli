package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		url        string
		wantRepo   string
		wantPrefix string
	}{
		{"https://github.com/org/repo", "repo", "https://"},
		{"https://github.com/org/repo.git", "repo", "https://"},
		{"git@github.com:org/repo.git", "repo", "git@"},
		{"git@gitlab.com:group/subgroup/repo", "repo", "git@"},
	}
	for _, tt := range tests {
		repoName, normalized, err := ParseGitURL(tt.url)
		if err != nil {
			t.Errorf("ParseGitURL(%q): %v", tt.url, err)
			continue
		}
		if repoName != tt.wantRepo {
			t.Errorf("ParseGitURL(%q) repo = %q, want %q", tt.url, repoName, tt.wantRepo)
		}
		if normalized == "" {
			t.Errorf("ParseGitURL(%q) normalized empty", tt.url)
		}
	}
}

func TestParseGitURL_Invalid(t *testing.T) {
	_, _, err := ParseGitURL("")
	if err == nil {
		t.Error("expected error for empty URL")
	}
	_, _, err = ParseGitURL("not-a-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestIsGitRepo(t *testing.T) {
	tmp := t.TempDir()
	if IsGitRepo(tmp) {
		t.Error("plain dir should not be git repo")
	}
	gitDir := filepath.Join(tmp, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if !IsGitRepo(tmp) {
		t.Error("dir with .git should be git repo")
	}
}

func TestCompareSemver(t *testing.T) {
	if compareSemver("v1.2.3", "v1.2.0") <= 0 {
		t.Error("v1.2.3 should be > v1.2.0")
	}
	if compareSemver("v2.0.0", "v1.9.9") <= 0 {
		t.Error("v2.0.0 should be > v1.9.9")
	}
	if compareSemver("v1.0.0", "v1.0.0") != 0 {
		t.Error("v1.0.0 should equal v1.0.0")
	}
}

func TestNewerTag(t *testing.T) {
	newer, isNewer := NewerTag("v1.0.0", "v1.1.0")
	if !isNewer || newer != "v1.1.0" {
		t.Errorf("NewerTag(v1.0.0, v1.1.0) = %q, %v; want v1.1.0, true", newer, isNewer)
	}
	newer, isNewer = NewerTag("v1.1.0", "v1.0.0")
	if isNewer || newer != "v1.1.0" {
		t.Errorf("NewerTag(v1.1.0, v1.0.0) = %q, %v; want v1.1.0, false", newer, isNewer)
	}
}
