package deps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	alib "mb/internal/shared/aliases"
)

// WriteMbcliYAMLAliasSection replaces the aliases section in mbcli.yaml from an in-memory map
// keyed by alib.StoreKey. Other top-level keys are preserved.
func WriteMbcliYAMLAliasSection(mbcliPath string, entries map[string]alib.Entry) error {
	root, err := readMbcliYAMLRootMap(mbcliPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		delete(root, "aliases")
	} else {
		aliasesMap, err := alib.AliasesYAMLMapFromEntries(entries)
		if err != nil {
			return err
		}
		root["aliases"] = aliasesMap
	}
	return writeMbcliYAMLRootAtomic(mbcliPath, root)
}

// UpsertMbcliYAMLAlias writes or updates one alias (name + entry.Env_vault slot) in mbcli.yaml.
// Other top-level keys are preserved when possible. Creates the file if missing.
// Rewriting may drop comments and reorder keys (yaml.Marshal).
func UpsertMbcliYAMLAlias(mbcliPath, name string, e alib.Entry) error {
	if err := alib.ValidateName(name); err != nil {
		return err
	}
	if err := alib.ValidateEntry(e); err != nil {
		return err
	}
	entries, err := ParseMbcliAliases(mbcliPath)
	if err != nil {
		return err
	}
	entries[alib.StoreKey(e.EnvVault, name)] = e
	return WriteMbcliYAMLAliasSection(mbcliPath, entries)
}

// RemoveMbcliYAMLAlias removes one alias identified by display name and env_vault (empty string = sem vault).
func RemoveMbcliYAMLAlias(mbcliPath, name, envVault string) error {
	if err := alib.ValidateName(name); err != nil {
		return err
	}
	entries, err := ParseMbcliAliases(mbcliPath)
	if err != nil {
		return err
	}
	sk := alib.StoreKey(envVault, name)
	if _, ok := entries[sk]; !ok {
		return fmt.Errorf("alias %q (vault %s) não existe em %q", name, formatVaultLabelWrite(envVault), mbcliPath)
	}
	delete(entries, sk)
	return WriteMbcliYAMLAliasSection(mbcliPath, entries)
}

func formatVaultLabelWrite(v string) string {
	if v == "" {
		return "(nenhum)"
	}
	return v
}

func readMbcliYAMLRootMap(mbcliPath string) (map[string]any, error) {
	data, err := os.ReadFile(mbcliPath)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mbcli.yaml: erro ao ler %q: %w", mbcliPath, err)
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("mbcli.yaml: %w", err)
	}
	if root == nil {
		return map[string]any{}, nil
	}
	return root, nil
}

func writeMbcliYAMLRootAtomic(mbcliPath string, root map[string]any) error {
	data, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("mbcli.yaml: erro ao serializar: %w", err)
	}
	dir := filepath.Dir(mbcliPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".mbcli.*.tmp")
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
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, mbcliPath)
}
