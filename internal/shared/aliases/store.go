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

const currentVersion = 2

// NormalizeFileAliasKeys rekeys legacy maps (display name as key) to StoreKey(env_vault, displayName).
func NormalizeFileAliasKeys(f *File) {
	if f == nil || f.Aliases == nil {
		return
	}
	norm := make(map[string]Entry, len(f.Aliases))
	for k, v := range f.Aliases {
		if _, _, ok := ParseStoreKey(k); ok {
			norm[k] = v
			continue
		}
		norm[StoreKey(v.EnvVault, k)] = v
	}
	f.Aliases = norm
}

// Load reads aliases.yaml; missing file yields an empty File (no error).
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &File{Version: currentVersion, Aliases: map[string]Entry{}}, nil
		}
		return nil, err
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse aliases: %w", err)
	}
	if root == nil {
		return &File{Version: currentVersion, Aliases: map[string]Entry{}}, nil
	}
	version := currentVersion
	switch v := root["version"].(type) {
	case int:
		version = v
	case int64:
		version = int(v)
	case float64:
		version = int(v)
	}
	raw := root["aliases"]
	if raw == nil {
		return &File{Version: version, Aliases: map[string]Entry{}}, nil
	}
	entries, err := ParseAliasesYAMLValue(raw)
	if err != nil {
		return nil, fmt.Errorf("parse aliases: %w", err)
	}
	return &File{Version: version, Aliases: entries}, nil
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
	NormalizeFileAliasKeys(f)
	aliasesDoc, err := AliasesYAMLMapFromEntries(f.Aliases)
	if err != nil {
		return err
	}
	root := map[string]any{
		"version": currentVersion,
		"aliases": aliasesDoc,
	}
	data, err := yaml.Marshal(root)
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

// Lookup resolves a single alias by display name only (legacy). Prefer ResolveForRunWithProject.
func Lookup(f *File, name string) (argv []string, envVault string, ok bool) {
	if f == nil || f.Aliases == nil {
		return nil, "", false
	}
	sk := StoreKey("", name)
	e, ok := f.Aliases[sk]
	if !ok || len(e.Command) == 0 {
		return nil, "", false
	}
	out := append([]string(nil), e.Command...)
	return out, e.EnvVault, true
}

// SortedStoreKeys returns lexicographic store keys for f.Aliases.
func SortedStoreKeys(f *File) []string {
	if f == nil || len(f.Aliases) == 0 {
		return nil
	}
	keys := make([]string, 0, len(f.Aliases))
	for k := range f.Aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// EffectiveEnvVault returns the vault to use for mb run: CLI wins over alias default.
func EffectiveEnvVault(cliVault, aliasVault string) string {
	if cliVault != "" {
		return cliVault
	}
	return aliasVault
}

// mergeProjectOverGlobal returns a new File: global entries first, then project entries
// overwrite on duplicate store keys (project wins).
func mergeProjectOverGlobal(global *File, project map[string]Entry) *File {
	out := &File{Version: currentVersion, Aliases: make(map[string]Entry)}
	if global != nil {
		out.Version = global.Version
		if out.Version == 0 {
			out.Version = currentVersion
		}
		for k, v := range global.Aliases {
			out.Aliases[k] = v
		}
	}
	for k, v := range project {
		out.Aliases[k] = v
	}
	return out
}

// ResolveForRunWithProject merges ~/.config/mb/aliases.yaml with project aliases (mbcli.yaml).
// cliVault is the MB global --env-vault (may be empty). When several aliases share the same
// display name, cliVault must be set to pick one or resolution returns an error.
func ResolveForRunWithProject(
	configDir string,
	project map[string]Entry,
	cliVault string,
	first string,
	rest []string,
) (argv []string, aliasVault string, resolved bool, err error) {
	path := FilePath(configDir)
	global, err := Load(path)
	if err != nil {
		return nil, "", false, err
	}
	merged := mergeProjectOverGlobal(global, project)
	argv0, v, ok, err := resolveMergedAlias(merged, cliVault, first)
	if err != nil {
		return nil, "", false, err
	}
	if !ok {
		return append([]string{first}, rest...), "", false, nil
	}
	out := append(argv0, rest...)
	return out, v, true, nil
}

func resolveMergedAlias(
	merged *File,
	cliVault, displayName string,
) (argv []string, envVault string, ok bool, err error) {
	if merged == nil || merged.Aliases == nil {
		return nil, "", false, nil
	}
	var hits []Entry
	for sk, e := range merged.Aliases {
		_, dn, pok := ParseStoreKey(sk)
		if !pok || dn != displayName {
			continue
		}
		if e.EnvVault != "" {
			_ = sk
		}
		hits = append(hits, e)
	}
	if len(hits) == 0 {
		return nil, "", false, nil
	}
	if len(hits) == 1 {
		e := hits[0]
		if len(e.Command) == 0 {
			return nil, "", false, nil
		}
		return append([]string(nil), e.Command...), e.EnvVault, true, nil
	}
	if cliVault == "" {
		return nil, "", false, fmt.Errorf(
			"vários aliases com o nome %q; indique --env-vault (ex.: mb --env-vault staging run %q)",
			displayName,
			displayName,
		)
	}
	var match *Entry
	for i := range hits {
		if hits[i].EnvVault == cliVault {
			if match != nil {
				return nil, "", false, fmt.Errorf(
					"vários aliases %q para o vault %q",
					displayName,
					cliVault,
				)
			}
			match = &hits[i]
		}
	}
	if match == nil {
		return nil, "", false, fmt.Errorf(
			"nenhum alias %q associado ao vault %q (use mb alias list para ver os vaults)",
			displayName, cliVault,
		)
	}
	if len(match.Command) == 0 {
		return nil, "", false, nil
	}
	return append([]string(nil), match.Command...), match.EnvVault, true, nil
}

// ResolveForRun applies MB alias expansion: first token may be an alias name; extra args append.
func ResolveForRun(
	configDir string,
	first string,
	rest []string,
) (argv []string, aliasVault string, resolved bool, err error) {
	return ResolveForRunWithProject(configDir, nil, "", first, rest)
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

// FindUniqueStoreKeyForDisplayName returns the sole store key for displayName, or n==0 / n>1.
func FindUniqueStoreKeyForDisplayName(
	m map[string]Entry,
	displayName string,
) (key string, e Entry, n int) {
	if m == nil {
		return "", Entry{}, 0
	}
	var keys []string
	for sk := range m {
		_, dn, ok := ParseStoreKey(sk)
		if ok && dn == displayName {
			keys = append(keys, sk)
		}
	}
	switch len(keys) {
	case 0:
		return "", Entry{}, 0
	case 1:
		return keys[0], m[keys[0]], 1
	default:
		return "", Entry{}, len(keys)
	}
}

// PickShellEntryForDisplayName returns the entry used for shell function generation, or ok false.
// Rule: if exactly one alias has this display name, use it; if several, only use the one with EnvVault == "";
// if several and none have empty vault, returns ok false and err non-nil.
func PickShellEntryForDisplayName(f *File, displayName string) (e Entry, ok bool, err error) {
	if f == nil || f.Aliases == nil {
		return Entry{}, false, nil
	}
	var hits []Entry
	for sk, ent := range f.Aliases {
		_, dn, pok := ParseStoreKey(sk)
		if !pok || dn != displayName {
			continue
		}
		hits = append(hits, ent)
	}
	switch len(hits) {
	case 0:
		return Entry{}, false, nil
	case 1:
		return hits[0], true, nil
	default:
		var def []Entry
		for _, h := range hits {
			if h.EnvVault == "" {
				def = append(def, h)
			}
		}
		if len(def) == 1 {
			return def[0], true, nil
		}
		if len(def) > 1 {
			return Entry{}, false, fmt.Errorf(
				"aliases: várias definições de %q sem env_vault",
				displayName,
			)
		}
		return Entry{}, false, fmt.Errorf(
			"aliases: várias definições de %q com env_vault; o shell só pode expor uma função por nome — "+
				"deixe uma delas sem vault ou renomeie",
			displayName,
		)
	}
}

// SortedDisplayNames returns sorted unique display names present in f.
func SortedDisplayNames(f *File) []string {
	if f == nil || len(f.Aliases) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	for sk := range f.Aliases {
		_, dn, ok := ParseStoreKey(sk)
		if !ok {
			continue
		}
		seen[dn] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// SortedNames is deprecated: returns sorted display names (not raw map keys).
func SortedNames(f *File) []string {
	return SortedDisplayNames(f)
}
