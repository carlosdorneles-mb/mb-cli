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

	for _, name := range []string{"all.sh", "log.sh", "memory.sh"} {
		full := filepath.Join(shellDir, name)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("file %s: %v", name, err)
		}
	}

	checksumPath := filepath.Join(shellDir, ".checksum")
	checksumData, err := os.ReadFile(checksumPath)
	if err != nil {
		t.Errorf(".checksum missing after first call: %v", err)
	}
	checksumStr := strings.TrimSpace(string(checksumData))
	if len(checksumStr) != 64 {
		t.Errorf(".checksum should be 64 hex chars, got %d: %q", len(checksumStr), checksumStr)
	}
	for _, c := range checksumStr {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		t.Errorf(".checksum contains non-hex char %q", c)
		break
	}

	allPath := filepath.Join(shellDir, "all.sh")
	allData, _ := os.ReadFile(allPath)
	if !strings.Contains(string(allData), "log.sh") {
		t.Errorf("all.sh should source log.sh, got: %s", allData)
	}
	if !strings.Contains(string(allData), "memory.sh") {
		t.Errorf("all.sh should source memory.sh, got: %s", allData)
	}
	logPath := filepath.Join(shellDir, "log.sh")
	logData, _ := os.ReadFile(logPath)
	if !strings.Contains(string(logData), "log()") {
		t.Errorf("log.sh should define log(), got: %s", logData)
	}
	memoryPath := filepath.Join(shellDir, "memory.sh")
	memoryData, _ := os.ReadFile(memoryPath)
	if !strings.Contains(string(memoryData), "mem_set()") {
		t.Errorf("memory.sh should define mem_set(), got: %s", memoryData)
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

func TestEnsureShellHelpers_overwritesWhenChecksumDiffers(t *testing.T) {
	configDir := t.TempDir()

	path, err := EnsureShellHelpers(configDir)
	if err != nil {
		t.Fatalf("EnsureShellHelpers: %v", err)
	}
	shellDir := filepath.Join(configDir, "lib", "shell")
	if path != shellDir {
		t.Fatalf("path = %q, want %q", path, shellDir)
	}

	// Corrompe .checksum para simular versão antiga
	checksumPath := filepath.Join(shellDir, ".checksum")
	if err := os.WriteFile(checksumPath, []byte("wrongchecksum\n"), 0o644); err != nil {
		t.Fatalf("write wrong checksum: %v", err)
	}
	// Corrompe all.sh para verificar que será sobrescrito
	allPath := filepath.Join(shellDir, "all.sh")
	if err := os.WriteFile(allPath, []byte("corrupted"), 0o644); err != nil {
		t.Fatalf("write corrupted all.sh: %v", err)
	}

	// Nova chamada deve reescrever tudo
	path2, err := EnsureShellHelpers(configDir)
	if err != nil {
		t.Fatalf("second EnsureShellHelpers: %v", err)
	}
	if path2 != shellDir {
		t.Errorf("path = %q, want %q", path2, shellDir)
	}

	allData, _ := os.ReadFile(allPath)
	if !strings.Contains(string(allData), "log.sh") {
		t.Errorf("all.sh should have been overwritten with embed content (sources log.sh), got: %s", allData)
	}
	if !strings.Contains(string(allData), "memory.sh") {
		t.Errorf("all.sh should have been overwritten with embed content (sources memory.sh), got: %s", allData)
	}
	checksumData, _ := os.ReadFile(checksumPath)
	checksumStr := strings.TrimSpace(string(checksumData))
	if len(checksumStr) != 64 {
		t.Errorf(".checksum should be updated to 64 hex chars, got %d: %q", len(checksumStr), checksumStr)
	}
}
