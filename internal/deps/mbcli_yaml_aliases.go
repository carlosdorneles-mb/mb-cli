package deps

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	alib "mb/internal/shared/aliases"
)

// ParseMbcliAliases reads mbcli.yaml and returns a map keyed by alib.StoreKey(env_vault, displayName).
// Missing file: empty map, nil error. Project aliases override global entries with the same store key at merge time.
func ParseMbcliAliases(mbcliPath string) (map[string]alib.Entry, error) {
	data, err := os.ReadFile(mbcliPath)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]alib.Entry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mbcli.yaml: erro ao ler %q: %w", mbcliPath, err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("mbcli.yaml: %w", err)
	}
	if root == nil {
		return map[string]alib.Entry{}, nil
	}
	raw, ok := root["aliases"]
	if !ok || raw == nil {
		return map[string]alib.Entry{}, nil
	}
	m, err := alib.ParseAliasesYAMLValue(raw)
	if err != nil {
		return nil, fmt.Errorf("mbcli.yaml: %w", err)
	}
	return m, nil
}
