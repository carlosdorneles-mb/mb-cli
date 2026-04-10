package deps

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"mb/internal/shared/envvault"
)

// MbcliEnvListEntry is one variable from mbcli.yaml envs for display (mb envs list).
type MbcliEnvListEntry struct {
	Vault, Key, Value string
}

// ParseMbcliProjectEnvs splits mbcli.yaml "envs" into the project default vault (root scalars)
// and named project vaults (keys whose value is a map of scalars).
// Missing file: empty maps, nil error. Invalid YAML or bad envs shape: error.
func ParseMbcliProjectEnvs(
	mbcliPath string,
) (defaultVars map[string]string, byVault map[string]map[string]string, err error) {
	defaultVars = map[string]string{}
	byVault = map[string]map[string]string{}

	data, err := os.ReadFile(mbcliPath)
	if errors.Is(err, os.ErrNotExist) {
		return defaultVars, byVault, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("mbcli.yaml: ler %q: %w", mbcliPath, err)
	}

	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, nil, fmt.Errorf("mbcli.yaml: %w", err)
	}

	raw, ok := root["envs"]
	if !ok || raw == nil {
		return defaultVars, byVault, nil
	}

	m, ok := envsAsStringAnyMap(raw)
	if !ok {
		return nil, nil, fmt.Errorf("mbcli.yaml: envs deve ser um mapa")
	}

	for k, v := range m {
		if inner, ok := asStringAnyMap(v); ok {
			if err := envvault.Validate(k); err != nil {
				return nil, nil, fmt.Errorf("mbcli.yaml envs.%s (vault): %w", k, err)
			}
			parsed := make(map[string]string)
			for ik, iv := range inner {
				s, serr := scalarEnvToString(iv)
				if serr != nil {
					return nil, nil, fmt.Errorf("mbcli.yaml envs.%s.%s: %w", k, ik, serr)
				}
				parsed[ik] = s
			}
			byVault[k] = parsed
			continue
		}
		s, serr := scalarEnvToString(v)
		if serr != nil {
			return nil, nil, fmt.Errorf("mbcli.yaml envs.%s: %w", k, serr)
		}
		defaultVars[k] = s
	}
	return defaultVars, byVault, nil
}

// LoadMbcliProjectEnvsForMerge returns env vars from mbcli.yaml for BuildEnvFileValues.
// Always merges root scalar keys (project default). If envVault is non-empty, merges
// envs[envVault] on top when that nested map exists.
func LoadMbcliProjectEnvsForMerge(mbcliPath, envVault string) (map[string]string, error) {
	def, byV, err := ParseMbcliProjectEnvs(mbcliPath)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(def)+8)
	for k, v := range def {
		out[k] = v
	}
	if envVault != "" {
		if inner, ok := byV[envVault]; ok {
			for k, v := range inner {
				out[k] = v
			}
		}
	}
	return out, nil
}

func projectMbcliVaultLabel(inner string) string {
	if inner == "default" {
		return "project"
	}
	return "project/" + inner
}

// mbcliListFilter returns how to filter MbcliProjectListEntries for listVault.
func mbcliListFilter(listVault string) (mode string, nested string, err error) {
	switch listVault {
	case "":
		return "all", "", nil
	case "default", "project":
		return "root", "", nil
	default:
		if strings.HasPrefix(listVault, "project/") {
			suf := strings.TrimPrefix(listVault, "project/")
			if suf == "" {
				return "", "", fmt.Errorf(
					`vault "project/" inválido: indique um nome após a barra (ex.: project/staging)`,
				)
			}
			if err := envvault.Validate(suf); err != nil {
				return "", "", fmt.Errorf("vault de projeto: %w", err)
			}
			return "projectNestedOnly", suf, nil
		}
		if err := envvault.Validate(listVault); err != nil {
			return "", "", err
		}
		return "merged", listVault, nil
	}
}

// MbcliProjectListEntries returns rows for mb envs list (storage column is projeto elsewhere).
// listVault "" lists all project vaults; "default" or "project" only root scalars;
// "project/<n>" only envs.<n> in the YAML (strict nested vault);
// plain "<n>" (e.g. with disk .env.<n>) uses merged root + envs.<n> to match runtime --env-vault.
// Vault column uses logical names "project" / "project/<n>".
func MbcliProjectListEntries(mbcliPath, listVault string) ([]MbcliEnvListEntry, error) {
	def, byV, err := ParseMbcliProjectEnvs(mbcliPath)
	if err != nil {
		return nil, err
	}
	mode, nested, err := mbcliListFilter(listVault)
	if err != nil {
		return nil, err
	}
	var out []MbcliEnvListEntry

	switch mode {
	case "all":
		for k, v := range def {
			out = append(
				out,
				MbcliEnvListEntry{Vault: projectMbcliVaultLabel("default"), Key: k, Value: v},
			)
		}
		for _, vn := range sortedStringKeys(byV) {
			for k, v := range byV[vn] {
				out = append(
					out,
					MbcliEnvListEntry{Vault: projectMbcliVaultLabel(vn), Key: k, Value: v},
				)
			}
		}
	case "root":
		for k, v := range def {
			out = append(
				out,
				MbcliEnvListEntry{Vault: projectMbcliVaultLabel("default"), Key: k, Value: v},
			)
		}
	case "merged":
		for k, v := range def {
			out = append(
				out,
				MbcliEnvListEntry{Vault: projectMbcliVaultLabel("default"), Key: k, Value: v},
			)
		}
		if inner, ok := byV[nested]; ok {
			for k, v := range inner {
				out = append(
					out,
					MbcliEnvListEntry{Vault: projectMbcliVaultLabel(nested), Key: k, Value: v},
				)
			}
		}
	case "projectNestedOnly":
		if inner, ok := byV[nested]; ok {
			for k, v := range inner {
				out = append(
					out,
					MbcliEnvListEntry{Vault: projectMbcliVaultLabel(nested), Key: k, Value: v},
				)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		vi, vj := out[i].Vault, out[j].Vault
		if vi != vj {
			return vi < vj
		}
		return out[i].Key < out[j].Key
	})
	return out, nil
}

func sortedStringKeys(m map[string]map[string]string) []string {
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

func scalarEnvToString(v any) (string, error) {
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
