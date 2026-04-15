package deps

import (
	"fmt"
	"strings"

	"mb/internal/shared/envvault"
)

// UpsertMbcliYAMLEnvs writes pairs into mbcli.yaml under envs.
// vault empty: scalars at envs root (logical project default).
// vault non-empty: envs.<vault>.<key> (nested map); vault must pass envvault.Validate.
// Preserves other top-level keys and other envs keys; rewrites may drop YAML comments.
func UpsertMbcliYAMLEnvs(mbcliPath, vault string, pairs map[string]string) error {
	if len(pairs) == 0 {
		return nil
	}
	if err := validateMbcliYAMLEnvVaultName(vault); err != nil {
		return err
	}
	if vault != "" {
		if err := envvault.Validate(vault); err != nil {
			return err
		}
	}
	for k := range pairs {
		if err := envvault.Validate(k); err != nil {
			return fmt.Errorf("mbcli.yaml envs: chave %q: %w", k, err)
		}
	}

	root, err := readMbcliYAMLRootMap(mbcliPath)
	if err != nil {
		return err
	}
	envsMap, err := ensureEnvsEditableMap(root)
	if err != nil {
		return err
	}

	if err := checkUpsertConflicts(envsMap, vault, pairs); err != nil {
		return err
	}

	if vault == "" {
		for k, v := range pairs {
			envsMap[k] = v
		}
	} else {
		inner, err := ensureNestedVaultMap(envsMap, vault)
		if err != nil {
			return err
		}
		for k, v := range pairs {
			inner[k] = v
		}
		envsMap[vault] = inner
	}

	root["envs"] = envsMap
	return writeMbcliYAMLRootAtomic(mbcliPath, root)
}

func ensureEnvsEditableMap(root map[string]any) (map[string]any, error) {
	raw, ok := root["envs"]
	if !ok || raw == nil {
		m := make(map[string]any)
		root["envs"] = m
		return m, nil
	}
	m, ok := envsAsStringAnyMap(raw)
	if !ok {
		return nil, fmt.Errorf("mbcli.yaml: envs deve ser um mapa")
	}
	return m, nil
}

func checkUpsertConflicts(envsMap map[string]any, vault string, pairs map[string]string) error {
	if vault == "" {
		for k := range pairs {
			raw, ok := envsMap[k]
			if !ok {
				continue
			}
			if _, isMap := asStringAnyMap(raw); isMap {
				return fmt.Errorf(
					"mbcli.yaml: envs.%q é um vault aninhado (mapa); defina variáveis dentro dele com --vault %q",
					k, k,
				)
			}
		}
		return nil
	}
	raw, ok := envsMap[vault]
	if !ok {
		return nil
	}
	if _, isMap := asStringAnyMap(raw); isMap {
		return nil
	}
	return fmt.Errorf(
		"mbcli.yaml: envs.%q é um valor escalar; não é possível usar --vault %q como submapa (remova ou renomeie a chave no YAML)",
		vault, vault,
	)
}

func ensureNestedVaultMap(envsMap map[string]any, vault string) (map[string]any, error) {
	raw, ok := envsMap[vault]
	if !ok {
		m := make(map[string]any)
		return m, nil
	}
	inner, ok := asStringAnyMap(raw)
	if !ok {
		return nil, fmt.Errorf(
			"mbcli.yaml: envs.%q é um valor escalar; não é possível gravar variáveis aninhadas (remova ou renomeie a chave no YAML)",
			vault,
		)
	}
	return inner, nil
}

// MbcliYAMLEnvKeysMissing returns keys that are not present in the mbcli.yaml envs scope (vault "" = root scalars).
func MbcliYAMLEnvKeysMissing(mbcliPath, vault string, keys []string) (missing []string, err error) {
	if err := validateMbcliYAMLEnvVaultName(vault); err != nil {
		return nil, err
	}
	if vault != "" {
		if err := envvault.Validate(vault); err != nil {
			return nil, err
		}
	}
	for _, k := range keys {
		if err := envvault.Validate(k); err != nil {
			return nil, fmt.Errorf("mbcli.yaml envs: chave %q: %w", k, err)
		}
	}
	root, err := readMbcliYAMLRootMap(mbcliPath)
	if err != nil {
		return nil, err
	}
	raw, ok := root["envs"]
	if !ok || raw == nil {
		return append([]string(nil), keys...), nil
	}
	envsMap, ok := envsAsStringAnyMap(raw)
	if !ok {
		return nil, fmt.Errorf("mbcli.yaml: envs deve ser um mapa")
	}
	if vault == "" {
		for _, key := range keys {
			v, ok := envsMap[key]
			if !ok {
				missing = append(missing, key)
				continue
			}
			if _, isMap := asStringAnyMap(v); isMap {
				return nil, fmt.Errorf(
					"mbcli.yaml: envs.%q é um vault aninhado (mapa); use mb envs unset --mbcli-yaml --vault %s",
					key, key,
				)
			}
		}
		return missing, nil
	}
	innerRaw, ok := envsMap[vault]
	if !ok {
		return append([]string(nil), keys...), nil
	}
	inner, ok := asStringAnyMap(innerRaw)
	if !ok {
		return nil, fmt.Errorf("mbcli.yaml: envs.%q não é um mapa de variáveis", vault)
	}
	for _, key := range keys {
		if _, ok := inner[key]; !ok {
			missing = append(missing, key)
		}
	}
	return missing, nil
}

// RemoveMbcliYAMLEnvKeys removes keys from mbcli.yaml envs for the given vault scope.
// vault "" removes root scalars only (errors if a key is a nested vault map).
// All keys must exist or no change is written; error lists missing keys.
func RemoveMbcliYAMLEnvKeys(mbcliPath, vault string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	missing, err := MbcliYAMLEnvKeysMissing(mbcliPath, vault, keys)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"variáveis inexistentes em mbcli.yaml (vault %s): %s",
			vaultLabelForErr(vault),
			strings.Join(missing, ", "),
		)
	}

	root, err := readMbcliYAMLRootMap(mbcliPath)
	if err != nil {
		return err
	}
	raw, ok := root["envs"]
	if !ok || raw == nil {
		return fmt.Errorf(
			"variáveis inexistentes em mbcli.yaml: %s",
			strings.Join(keys, ", "),
		)
	}
	envsMap, ok := envsAsStringAnyMap(raw)
	if !ok {
		return fmt.Errorf("mbcli.yaml: envs deve ser um mapa")
	}

	if vault == "" {
		for _, key := range keys {
			delete(envsMap, key)
		}
	} else {
		innerRaw := envsMap[vault]
		inner, _ := asStringAnyMap(innerRaw)
		for _, key := range keys {
			delete(inner, key)
		}
		if len(inner) == 0 {
			delete(envsMap, vault)
		} else {
			envsMap[vault] = inner
		}
	}

	if len(envsMap) == 0 {
		delete(root, "envs")
	} else {
		root["envs"] = envsMap
	}
	return writeMbcliYAMLRootAtomic(mbcliPath, root)
}

func vaultLabelForErr(vault string) string {
	if vault == "" {
		return "raiz (project)"
	}
	return vault
}

// validateMbcliYAMLEnvVaultName rejects logical list vault names that are not YAML paths.
func validateMbcliYAMLEnvVaultName(vault string) error {
	if vault == "" {
		return nil
	}
	if vault == "project" || strings.HasPrefix(vault, "project/") {
		return fmt.Errorf(
			`mbcli.yaml: para variáveis na raiz de envs omita --vault; "project" e "project/*" são nomes reservados em mb envs list`,
		)
	}
	return nil
}
