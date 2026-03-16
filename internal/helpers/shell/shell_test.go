package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureShellHelpers(t *testing.T) {
	configDir := t.TempDir()

	path, err := EnsureShellHelpers(configDir)
	if err != nil {
		t.Fatalf("EnsureShellHelpers: %v", err)
	}
	if path == "" {
		t.Fatal("EnsureShellHelpers returned empty path")
	}

	indexPath := filepath.Join(configDir, "lib", "shell", "index.sh")
	if path != indexPath {
		t.Errorf("path = %q, want %q", path, indexPath)
	}

	for _, name := range []string{"index.sh", "log.sh"} {
		full := filepath.Join(configDir, "lib", "shell", name)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("file %s: %v", name, err)
		}
	}

	indexData, _ := os.ReadFile(indexPath)
	if !strings.Contains(string(indexData), "log.sh") {
		t.Errorf("index.sh should source log.sh, got: %s", indexData)
	}
	logPath := filepath.Join(configDir, "lib", "shell", "log.sh")
	logData, _ := os.ReadFile(logPath)
	if !strings.Contains(string(logData), "log()") {
		t.Errorf("log.sh should define log(), got: %s", logData)
	}

	// Idempotent: second call returns same path
	path2, err2 := EnsureShellHelpers(configDir)
	if err2 != nil {
		t.Fatalf("second EnsureShellHelpers: %v", err2)
	}
	if path2 != path {
		t.Errorf("second call path = %q, want %q", path2, path)
	}
}
