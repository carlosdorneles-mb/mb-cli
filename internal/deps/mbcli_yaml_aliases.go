package deps

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	alib "mb/internal/shared/aliases"
	"mb/internal/shared/envvault"
)

// ParseMbcliAliases reads mbcli.yaml and returns a flat map of alias name -> entry.
// Missing file: empty map, nil error. Project aliases override global names at merge time.
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
	return parseAliasesYAMLValue(raw)
}

func parseAliasesYAMLValue(raw any) (map[string]alib.Entry, error) {
	m, ok := envsAsStringAnyMap(raw)
	if !ok {
		return nil, fmt.Errorf("mbcli.yaml: aliases precisa ser um mapa")
	}
	out := make(map[string]alib.Entry)
	for k, v := range m {
		if err := validateMbcliAliasTopLevelKey(k); err != nil {
			return nil, fmt.Errorf("mbcli.yaml aliases.%s: %w", k, err)
		}
		inner, ok := asStringAnyMap(v)
		if !ok {
			return nil, fmt.Errorf("mbcli.yaml aliases.%s: esperava-se um mapa", k)
		}
		if hasCommandKey(inner) {
			e, err := decodeAliasEntry(inner, "", k)
			if err != nil {
				return nil, fmt.Errorf("mbcli.yaml aliases.%s: %w", k, err)
			}
			if err := mergeAliasUnique(out, k, e); err != nil {
				return nil, err
			}
			continue
		}
		if err := validateMbcliAliasNestedGroupName(k); err != nil {
			return nil, fmt.Errorf("mbcli.yaml aliases.%s (grupo): %w", k, err)
		}
		for ik, iv := range inner {
			sub, ok := asStringAnyMap(iv)
			if !ok {
				return nil, fmt.Errorf("mbcli.yaml aliases.%s.%s: esperava-se um mapa com command", k, ik)
			}
			e, err := decodeAliasEntry(sub, k, ik)
			if err != nil {
				return nil, fmt.Errorf("mbcli.yaml aliases.%s.%s: %w", k, ik, err)
			}
			if err := mergeAliasUnique(out, ik, e); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func hasCommandKey(m map[string]any) bool {
	_, ok := m["command"]
	return ok
}

func decodeAliasEntry(m map[string]any, implicitVault, aliasName string) (alib.Entry, error) {
	cmd, err := stringSliceFromAny(m["command"], fmt.Sprintf("aliases.%s.command", aliasName))
	if err != nil {
		return alib.Entry{}, err
	}
	_, vaultExplicit := m["env_vault"]
	vaultStr, _ := scalarEnvToString(m["env_vault"])
	var envVault string
	switch {
	case vaultExplicit:
		envVault = vaultStr
	case implicitVault != "":
		envVault = implicitVault
	}
	e := alib.Entry{Command: cmd, EnvVault: envVault}
	if err := alib.ValidateName(aliasName); err != nil {
		return alib.Entry{}, err
	}
	if err := alib.ValidateEntry(e); err != nil {
		return alib.Entry{}, err
	}
	return e, nil
}

func mergeAliasUnique(dst map[string]alib.Entry, name string, e alib.Entry) error {
	if _, dup := dst[name]; dup {
		return fmt.Errorf("mbcli.yaml: alias %q duplicado (os nomes precisam ser únicos após expandir os grupos)", name)
	}
	dst[name] = e
	return nil
}

func stringSliceFromAny(raw any, ctx string) ([]string, error) {
	if raw == nil {
		return nil, fmt.Errorf("%s: command ausente", ctx)
	}
	switch t := raw.(type) {
	case []string:
		if len(t) == 0 {
			return nil, fmt.Errorf("%s: command vazio", ctx)
		}
		return append([]string(nil), t...), nil
	case []any:
		if len(t) == 0 {
			return nil, fmt.Errorf("%s: command vazio", ctx)
		}
		out := make([]string, 0, len(t))
		for i, v := range t {
			s, err := scalarEnvToString(v)
			if err != nil {
				return nil, fmt.Errorf("%s[%d]: %w", ctx, i, err)
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s: command precisa ser uma lista", ctx)
	}
}

func validateMbcliAliasTopLevelKey(k string) error {
	if strings.TrimSpace(k) == "" {
		return fmt.Errorf("chave vazia")
	}
	if k == "project" || strings.HasPrefix(k, "project/") {
		return fmt.Errorf(`chave "project" e prefixo "project/" são reservados`)
	}
	return nil
}

func validateMbcliAliasNestedGroupName(k string) error {
	if k == "project" {
		return fmt.Errorf(`nome "project" é reservado`)
	}
	if strings.HasPrefix(k, "project/") {
		return fmt.Errorf(`prefixo "project/" é reservado`)
	}
	return envvault.Validate(k)
}
