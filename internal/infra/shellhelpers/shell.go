package shellhelpers

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed *.sh
var shellFS embed.FS

const checksumFile = ".checksum"

// embeddedShellFiles returns the names of embedded .sh files, sorted for deterministic checksum.
func embeddedShellFiles() ([]string, error) {
	entries, err := fs.Glob(shellFS, "*.sh")
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	return entries, nil
}

// helpersChecksum returns SHA256 (hex) of the concatenation of the contents of embedded .sh files.
func helpersChecksum() (string, error) {
	files, err := embeddedShellFiles()
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, name := range files {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnsureShellHelpers creates the lib/shell directory under configDir (if it does not exist),
// writes all embedded .sh files (automatically discovered), and returns the absolute path
// of the lib/shell directory. This path is passed to the plugin as MB_HELPERS_PATH.
//
// On every invocation (for example mb plugins sync), all embedded .sh files are rewritten
// on disk to match the current binary. Any .sh files under lib/shell that are no longer
// part of the embed are removed. The .checksum file is updated with the current aggregate
// hash (useful for inspection or external tooling).
func EnsureShellHelpers(configDir string) (string, error) {
	shellDir := filepath.Join(configDir, "lib", "shell")
	files, err := embeddedShellFiles()
	if err != nil {
		return "", err
	}
	currentChecksum, err := helpersChecksum()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(shellDir, 0o755); err != nil {
		return "", err
	}

	embedded := make(map[string]struct{}, len(files))
	for _, name := range files {
		embedded[name] = struct{}{}
	}

	for _, name := range files {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(shellDir, name), data, 0o644); err != nil {
			return "", err
		}
	}

	entries, err := os.ReadDir(shellDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == checksumFile {
			continue
		}
		if !strings.HasSuffix(name, ".sh") {
			continue
		}
		if _, ok := embedded[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(shellDir, name)); err != nil {
			return "", err
		}
	}

	checksumPath := filepath.Join(shellDir, checksumFile)
	if err := os.WriteFile(checksumPath, []byte(currentChecksum+"\n"), 0o644); err != nil {
		return "", err
	}
	return filepath.Abs(shellDir)
}
