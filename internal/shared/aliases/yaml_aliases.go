package aliases

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"mb/internal/shared/envvault"
)

// ParseAliasesYAMLValue parses the YAML value under "aliases" (map shape) into a map
// keyed by StoreKey(env_vault, displayName). Used for ~/.config/mb/aliases.yaml and
// the same grammar applies to mbcli.yaml aliases.
func ParseAliasesYAMLValue(raw any) (map[string]Entry, error) {
	m, ok := envsAsStringAnyMap(raw)
	if !ok {
		return nil, fmt.Errorf("aliases precisa ser um mapa")
	}
	out := make(map[string]Entry)
	for _, k := range sortedStringKeysAny(m) {
		v := m[k]
		if err := validateAliasYAMLTopLevelKey(k); err != nil {
			return nil, fmt.Errorf("aliases.%s: %w", k, err)
		}
		if isYAMLSequence(v) {
			cmd, err := stringSliceFromAny(v, fmt.Sprintf("aliases.%s", k))
			if err != nil {
				return nil, fmt.Errorf("aliases.%s: %w", k, err)
			}
			e := Entry{Command: cmd}
			if err := ValidateName(k); err != nil {
				return nil, fmt.Errorf("aliases.%s: %w", k, err)
			}
			if err := ValidateEntry(e); err != nil {
				return nil, fmt.Errorf("aliases.%s: %w", k, err)
			}
			if err := mergeStoreUnique(out, k, e); err != nil {
				return nil, err
			}
			continue
		}
		inner, ok := asStringAnyMap(v)
		if !ok {
			return nil, fmt.Errorf("aliases.%s: esperava-se uma lista de comando ou um mapa", k)
		}
		if hasCommandKey(inner) {
			e, err := decodeAliasEntryMap(inner, "", k)
			if err != nil {
				return nil, fmt.Errorf("aliases.%s: %w", k, err)
			}
			if err := mergeStoreUnique(out, k, e); err != nil {
				return nil, err
			}
			continue
		}
		if err := validateAliasYAMLNestedGroupName(k); err != nil {
			return nil, fmt.Errorf("aliases.%s (grupo): %w", k, err)
		}
		for _, ik := range sortedStringKeysAny(inner) {
			iv := inner[ik]
			if isYAMLSequence(iv) {
				cmd, err := stringSliceFromAny(iv, fmt.Sprintf("aliases.%s.%s", k, ik))
				if err != nil {
					return nil, fmt.Errorf("aliases.%s.%s: %w", k, ik, err)
				}
				e := Entry{Command: cmd, EnvVault: k}
				if err := ValidateName(ik); err != nil {
					return nil, fmt.Errorf("aliases.%s.%s: %w", k, ik, err)
				}
				if err := ValidateEntry(e); err != nil {
					return nil, fmt.Errorf("aliases.%s.%s: %w", k, ik, err)
				}
				if err := mergeStoreUnique(out, ik, e); err != nil {
					return nil, err
				}
				continue
			}
			sub, ok := asStringAnyMap(iv)
			if !ok {
				return nil, fmt.Errorf(
					"aliases.%s.%s: esperava-se uma lista de comando ou um mapa com command",
					k,
					ik,
				)
			}
			e, err := decodeAliasEntryMap(sub, k, ik)
			if err != nil {
				return nil, fmt.Errorf("aliases.%s.%s: %w", k, ik, err)
			}
			if err := mergeStoreUnique(out, ik, e); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func mergeStoreUnique(dst map[string]Entry, displayName string, e Entry) error {
	sk := StoreKey(e.EnvVault, displayName)
	if _, dup := dst[sk]; dup {
		return fmt.Errorf(
			"alias %q duplicado para o mesmo vault %q",
			displayName,
			formatVaultForErr(e.EnvVault),
		)
	}
	dst[sk] = e
	return nil
}

func formatVaultForErr(v string) string {
	if v == "" {
		return "(nenhum)"
	}
	return v
}

func hasCommandKey(m map[string]any) bool {
	_, ok := m["command"]
	return ok
}

func decodeAliasEntryMap(m map[string]any, implicitVault, aliasName string) (Entry, error) {
	cmd, err := stringSliceFromAny(m["command"], fmt.Sprintf("aliases.%s.command", aliasName))
	if err != nil {
		return Entry{}, err
	}
	_, vaultExplicit := m["env_vault"]
	vaultStr, _ := scalarStringFromYAML(m["env_vault"])
	var envVault string
	switch {
	case vaultExplicit:
		envVault = vaultStr
	case implicitVault != "":
		envVault = implicitVault
	}
	e := Entry{Command: cmd, EnvVault: envVault}
	if err := ValidateName(aliasName); err != nil {
		return Entry{}, err
	}
	if err := ValidateEntry(e); err != nil {
		return Entry{}, err
	}
	return e, nil
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
			s, err := scalarStringFromYAML(v)
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

func isYAMLSequence(v any) bool {
	switch v.(type) {
	case []any, []string:
		return true
	default:
		return false
	}
}

func sortedStringKeysAny(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func envsAsStringAnyMap(raw any) (map[string]any, bool) {
	if m, ok := raw.(map[string]any); ok {
		return m, true
	}
	if m2, ok := raw.(map[any]any); ok {
		out := make(map[string]any, len(m2))
		for k, v := range m2 {
			ks, ok := k.(string)
			if !ok {
				return nil, false
			}
			out[ks] = v
		}
		return out, true
	}
	return nil, false
}

func asStringAnyMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	if m2, ok := v.(map[any]any); ok {
		out := make(map[string]any, len(m2))
		for k, val := range m2 {
			ks, ok := k.(string)
			if !ok {
				return nil, false
			}
			out[ks] = val
		}
		return out, true
	}
	return nil, false
}

func scalarStringFromYAML(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	switch t := v.(type) {
	case string:
		return t, nil
	case bool:
		if t {
			return "true", nil
		}
		return "false", nil
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(rv.Int(), 10), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(rv.Uint(), 10), nil
		case reflect.Map, reflect.Slice, reflect.Array:
			return "", fmt.Errorf("esperado valor escalar, obteve %s", rv.Kind().String())
		default:
			return fmt.Sprint(t), nil
		}
	}
}

func validateAliasYAMLTopLevelKey(k string) error {
	if strings.TrimSpace(k) == "" {
		return fmt.Errorf("chave vazia")
	}
	if k == "project" || strings.HasPrefix(k, "project/") {
		return fmt.Errorf(`chave "project" e prefixo "project/" são reservados`)
	}
	return nil
}

func validateAliasYAMLNestedGroupName(k string) error {
	if k == "project" {
		return fmt.Errorf(`nome "project" é reservado`)
	}
	if strings.HasPrefix(k, "project/") {
		return fmt.Errorf(`prefixo "project/" é reservado`)
	}
	return envvault.Validate(k)
}

// AliasesYAMLMapFromEntries builds the nested shorthand aliases map for YAML serialization
// (root lists for env_vault "", groups for non-empty vault). Keys in entries must be StoreKeys.
func AliasesYAMLMapFromEntries(entries map[string]Entry) (map[string]any, error) {
	vaultUsed := make(map[string]struct{})
	for _, e := range entries {
		if e.EnvVault != "" {
			vaultUsed[e.EnvVault] = struct{}{}
		}
	}
	for sk, e := range entries {
		vault, displayName, ok := ParseStoreKey(sk)
		if !ok {
			return nil, fmt.Errorf("aliases: chave interna inválida %q", sk)
		}
		if e.EnvVault != vault {
			return nil, fmt.Errorf("aliases: inconsistência vault na chave %q", sk)
		}
		if e.EnvVault == "" {
			if _, clash := vaultUsed[displayName]; clash {
				return nil, fmt.Errorf(
					"aliases: nome %q sem env_vault conflita com o vault %q usado por outro(s) alias(es); renomeie o alias ou o vault",
					displayName,
					displayName,
				)
			}
		}
	}

	out := make(map[string]any)
	var rootNames []string
	for sk, e := range entries {
		vault, displayName, ok := ParseStoreKey(sk)
		if !ok || vault != "" {
			continue
		}
		_ = e
		rootNames = append(rootNames, displayName)
	}
	sort.Strings(rootNames)
	for _, displayName := range rootNames {
		sk := StoreKey("", displayName)
		out[displayName] = commandToYAMLAnySlice(entries[sk].Command)
	}

	byVault := make(map[string][]string)
	for sk, e := range entries {
		vault, displayName, ok := ParseStoreKey(sk)
		if !ok || vault == "" {
			continue
		}
		if e.EnvVault != vault {
			continue
		}
		byVault[vault] = append(byVault[vault], displayName)
	}
	vaultKeys := make([]string, 0, len(byVault))
	for v := range byVault {
		vaultKeys = append(vaultKeys, v)
	}
	sort.Strings(vaultKeys)
	for _, v := range vaultKeys {
		names := byVault[v]
		sort.Strings(names)
		inner := make(map[string]any, len(names))
		for _, displayName := range names {
			sk := StoreKey(v, displayName)
			inner[displayName] = commandToYAMLAnySlice(entries[sk].Command)
		}
		out[v] = inner
	}
	return out, nil
}

func commandToYAMLAnySlice(cmd []string) []any {
	out := make([]any, len(cmd))
	for i, s := range cmd {
		out[i] = s
	}
	return out
}
