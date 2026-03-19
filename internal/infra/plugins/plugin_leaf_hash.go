package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/shared/safepath"
)

// PluginLeafConfigHash returns a deterministic SHA-256 hex digest for a plugin leaf directory.
// It combines the raw manifest bytes with SHA-256 of each referenced file (entrypoint, flag
// entrypoints, env_files paths, readme), in a canonical sorted block, then hashes that block.
func PluginLeafConfigHash(raw []byte, manifest *Manifest, baseDir string) (string, error) {
	relPaths := collectLeafRefPaths(manifest)
	seen := make(map[string]struct{}, len(relPaths))
	for _, p := range relPaths {
		seen[p] = struct{}{}
	}
	sorted := make([]string, 0, len(seen))
	for p := range seen {
		sorted = append(sorted, p)
	}
	sort.Strings(sorted)

	lines := make([]string, 0, len(sorted)+1)
	mh := sha256.Sum256(raw)
	lines = append(lines, "manifest.yaml\t"+hex.EncodeToString(mh[:]))

	for _, rel := range sorted {
		abs := filepath.Join(baseDir, rel)
		if err := safepath.ValidateUnderDir(abs, baseDir); err != nil {
			return "", fmt.Errorf("ref %q: %w", rel, err)
		}
		b, err := os.ReadFile(abs)
		if err != nil {
			return "", fmt.Errorf("read ref %q: %w", rel, err)
		}
		sum := sha256.Sum256(b)
		lines = append(lines, rel+"\t"+hex.EncodeToString(sum[:]))
	}
	sort.Strings(lines)

	block := strings.Join(lines, "\n")
	final := sha256.Sum256([]byte(block))
	return hex.EncodeToString(final[:]), nil
}

func collectLeafRefPaths(m *Manifest) []string {
	var out []string
	if m.Entrypoint != "" {
		out = append(out, filepath.ToSlash(m.Entrypoint))
	}
	for _, e := range m.Flags.List {
		if e.Entrypoint != "" {
			out = append(out, filepath.ToSlash(e.Entrypoint))
		}
	}
	for _, e := range m.EnvFiles.List {
		if e.File != "" {
			out = append(out, filepath.ToSlash(e.File))
		}
	}
	if m.Readme != "" {
		out = append(out, filepath.ToSlash(m.Readme))
	}
	return out
}
