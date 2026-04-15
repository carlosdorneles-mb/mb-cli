package aliases

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"mb/internal/shared/envvault"
)

const currentVersion = 1

// Load reads aliases.yaml; missing file yields an empty File (no error).
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &File{Version: currentVersion, Aliases: map[string]Entry{}}, nil
		}
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse aliases: %w", err)
	}
	if f.Aliases == nil {
		f.Aliases = map[string]Entry{}
	}
	if f.Version == 0 {
		f.Version = currentVersion
	}
	return &f, nil
}

// Save writes aliases atomically with restrictive permissions.
func Save(path string, f *File) error {
	if f == nil {
		return errors.New("aliases: arquivo nil")
	}
	if f.Version == 0 {
		f.Version = currentVersion
	}
	if f.Aliases == nil {
		f.Aliases = map[string]Entry{}
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return atomicWrite(path, append([]byte{}, data...))
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".aliases.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// Lookup returns argv for name and whether it exists.
func Lookup(f *File, name string) (argv []string, envVault string, ok bool) {
	if f == nil || f.Aliases == nil {
		return nil, "", false
	}
	e, ok := f.Aliases[name]
	if !ok || len(e.Command) == 0 {
		return nil, "", false
	}
	out := append([]string(nil), e.Command...)
	return out, e.EnvVault, true
}

// SortedNames returns alias names in lexicographic order.
func SortedNames(f *File) []string {
	if f == nil || len(f.Aliases) == 0 {
		return nil
	}
	names := make([]string, 0, len(f.Aliases))
	for k := range f.Aliases {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// EffectiveEnvVault returns the vault to use for mb run: CLI wins over alias default.
func EffectiveEnvVault(cliVault, aliasVault string) string {
	if cliVault != "" {
		return cliVault
	}
	return aliasVault
}

// ResolveForRun applies MB alias expansion: first token may be an alias name; extra args append.
func ResolveForRun(
	configDir string,
	first string,
	rest []string,
) (argv []string, aliasVault string, resolved bool, err error) {
	path := FilePath(configDir)
	f, err := Load(path)
	if err != nil {
		return nil, "", false, err
	}
	argv0, v, ok := Lookup(f, first)
	if !ok {
		return append([]string{first}, rest...), "", false, nil
	}
	out := append(argv0, rest...)
	return out, v, true, nil
}

// ValidateEntry checks command and optional vault.
func ValidateEntry(e Entry) error {
	if len(e.Command) == 0 {
		return errors.New("comando do alias vazio")
	}
	for _, a := range e.Command {
		if a == "" {
			return errors.New("comando do alias contém argumento vazio")
		}
	}
	if e.EnvVault != "" {
		if err := envvault.ValidateConfigurableVault(e.EnvVault); err != nil {
			return err
		}
	}
	return nil
}
