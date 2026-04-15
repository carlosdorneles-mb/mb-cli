package deps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	alib "mb/internal/shared/aliases"
)

// UpsertMbcliYAMLAlias writes or updates a flat alias under aliases in mbcli.yaml.
// Other top-level keys are preserved when possible. Creates the file if missing.
// Rewriting may drop comments and reorder keys (yaml.Marshal).
func UpsertMbcliYAMLAlias(mbcliPath, name string, e alib.Entry) error {
	if err := alib.ValidateName(name); err != nil {
		return err
	}
	if err := alib.ValidateEntry(e); err != nil {
		return err
	}
	root, err := readMbcliYAMLRootMap(mbcliPath)
	if err != nil {
		return err
	}
	entries, err := ParseMbcliAliases(mbcliPath)
	if err != nil {
		return err
	}
	entries[name] = e
	root["aliases"] = aliasesEntriesToYAMLMap(entries)
	return writeMbcliYAMLRootAtomic(mbcliPath, root)
}

// RemoveMbcliYAMLAlias removes a flat alias name from mbcli.yaml aliases.
func RemoveMbcliYAMLAlias(mbcliPath, name string) error {
	if err := alib.ValidateName(name); err != nil {
		return err
	}
	root, err := readMbcliYAMLRootMap(mbcliPath)
	if err != nil {
		return err
	}
	entries, err := ParseMbcliAliases(mbcliPath)
	if err != nil {
		return err
	}
	if _, ok := entries[name]; !ok {
		return fmt.Errorf("alias %q não existe em %q", name, mbcliPath)
	}
	delete(entries, name)
	if len(entries) == 0 {
		delete(root, "aliases")
	} else {
		root["aliases"] = aliasesEntriesToYAMLMap(entries)
	}
	return writeMbcliYAMLRootAtomic(mbcliPath, root)
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

func aliasesEntriesToYAMLMap(entries map[string]alib.Entry) map[string]any {
	out := make(map[string]any, len(entries))
	for k, e := range entries {
		cmd := make([]any, len(e.Command))
		for i, s := range e.Command {
			cmd[i] = s
		}
		m := map[string]any{"command": cmd}
		if e.EnvVault != "" {
			m["env_vault"] = e.EnvVault
		}
		out[k] = m
	}
	return out
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
