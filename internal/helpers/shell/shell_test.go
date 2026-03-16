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

	shellDir := filepath.Join(configDir, "lib", "shell")
	if path != shellDir {
		t.Errorf("path = %q, want directory %q", path, shellDir)
	}

	for _, name := range []string{"all.sh", "log.sh"} {
		full := filepath.Join(shellDir, name)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("file %s: %v", name, err)
		}
	}

	allPath := filepath.Join(shellDir, "all.sh")
	allData, _ := os.ReadFile(allPath)
	if !strings.Contains(string(allData), "log.sh") {
		t.Errorf("all.sh should source log.sh, got: %s", allData)
	}
	logPath := filepath.Join(shellDir, "log.sh")
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
