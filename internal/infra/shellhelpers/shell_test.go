package shellhelpers

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

	for _, name := range []string{"all.sh", "log.sh", "memory.sh", "string.sh", "kubernetes.sh"} {
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
	if !strings.Contains(string(allData), "string.sh") {
		t.Errorf("all.sh should source string.sh, got: %s", allData)
	}
	if !strings.Contains(string(allData), "kubernetes.sh") {
		t.Errorf("all.sh should source kubernetes.sh, got: %s", allData)
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
	stringPath := filepath.Join(shellDir, "string.sh")
	stringData, _ := os.ReadFile(stringPath)
	if !strings.Contains(string(stringData), "str_to_upper()") {
		t.Errorf("string.sh should define str_to_upper(), got: %s", stringData)
	}
	kubernetesPath := filepath.Join(shellDir, "kubernetes.sh")
	kubernetesData, _ := os.ReadFile(kubernetesPath)
	if !strings.Contains(string(kubernetesData), "kb_check_installed()") {
		t.Errorf("kubernetes.sh should define kb_check_installed(), got: %s", kubernetesData)
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
		t.Errorf(
			"all.sh should have been overwritten with embed content (sources log.sh), got: %s",
			allData,
		)
	}
	if !strings.Contains(string(allData), "memory.sh") {
		t.Errorf(
			"all.sh should have been overwritten with embed content (sources memory.sh), got: %s",
			allData,
		)
	}
	if !strings.Contains(string(allData), "string.sh") {
		t.Errorf(
			"all.sh should have been overwritten with embed content (sources string.sh), got: %s",
			allData,
		)
	}
	if !strings.Contains(string(allData), "kubernetes.sh") {
		t.Errorf(
			"all.sh should have been overwritten with embed content (sources kubernetes.sh), got: %s",
			allData,
		)
	}
	checksumData, _ := os.ReadFile(checksumPath)
	checksumStr := strings.TrimSpace(string(checksumData))
	if len(checksumStr) != 64 {
		t.Errorf(
			".checksum should be updated to 64 hex chars, got %d: %q",
			len(checksumStr),
			checksumStr,
		)
	}
}

func TestEnsureShellHelpers_secondCallRestoresAllShWhenChecksumStillCorrect(t *testing.T) {
	configDir := t.TempDir()
	shellDir := filepath.Join(configDir, "lib", "shell")

	if _, err := EnsureShellHelpers(configDir); err != nil {
		t.Fatalf("first: %v", err)
	}
	allPath := filepath.Join(shellDir, "all.sh")
	if err := os.WriteFile(
		allPath,
		[]byte("# stale — embed should win on next sync"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureShellHelpers(configDir); err != nil {
		t.Fatalf("second: %v", err)
	}
	if data, _ := os.ReadFile(allPath); !strings.Contains(string(data), "log.sh") {
		t.Errorf(
			"second call should restore all.sh from embed even when .checksum unchanged, got %q",
			data,
		)
	}
}

func TestEnsureShellHelpers_removesOrphanSh(t *testing.T) {
	configDir := t.TempDir()
	shellDir := filepath.Join(configDir, "lib", "shell")

	if _, err := EnsureShellHelpers(configDir); err != nil {
		t.Fatalf("first: %v", err)
	}
	orphan := filepath.Join(shellDir, "orphan_not_in_embed.sh")
	if err := os.WriteFile(orphan, []byte("# orphan"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureShellHelpers(configDir); err != nil {
		t.Fatalf("second: %v", err)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Errorf("orphan .sh should be removed, still exists")
	}
}
