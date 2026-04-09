package envs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/deps"
	"mb/internal/ports"
	"mb/internal/shared/envgroup"
)

const secretsSuffix = ".secrets"

// ListRow is one variable shown by mb envs list.
type ListRow struct {
	Key, Value, Group string
}

// CollectListRows returns rows for default and all valid group env files, optionally revealing secrets.
func CollectListRows(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
	listGroup string,
	showSecrets bool,
) ([]ListRow, error) {
	if listGroup != "" {
		if err := envgroup.Validate(listGroup); err != nil {
			return nil, err
		}
		p, err := envgroup.FilePath(paths.ConfigDir, listGroup)
		if err != nil {
			return nil, err
		}
		return rowsForPath(secrets, onePassword, p, listGroup, showSecrets)
	}

	defRows, err := rowsForPath(secrets, onePassword, paths.DefaultEnvPath, "default", showSecrets)
	if err != nil {
		return nil, err
	}
	rows := append([]ListRow(nil), defRows...)
	matches, err := filepath.Glob(filepath.Join(paths.ConfigDir, ".env.*"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	for _, path := range matches {
		base := filepath.Base(path)
		if !strings.HasPrefix(base, ".env.") {
			continue
		}
		g := strings.TrimPrefix(base, ".env.")
		if g == "" || envgroup.Validate(g) != nil {
			continue
		}
		if strings.HasSuffix(path, secretsSuffix) {
			continue
		}
		r, err := rowsForPath(secrets, onePassword, path, g, showSecrets)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	sort.Slice(rows, func(i, j int) bool {
		gi, gj := rows[i].Group, rows[j].Group
		if gi != gj {
			if gi == "default" {
				return true
			}
			if gj == "default" {
				return false
			}
			return gi < gj
		}
		if rows[i].Key != rows[j].Key {
			return rows[i].Key < rows[j].Key
		}
		return rows[i].Value < rows[j].Value
	})
	return rows, nil
}

func rowsForPath(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	path, group string,
	showSecrets bool,
) ([]ListRow, error) {
	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return nil, err
	}
	secretKeys, err := deps.LoadSecretKeys(path)
	if err != nil {
		return nil, err
	}
	keyringGroup := KeyringGroup(group)
	seen := make(map[string]bool)
	var rows []ListRow
	for _, key := range sortedMapKeys(values) {
		seen[key] = true
		if isSecretKey(secretKeys, key) {
			val := "***"
			if showSecrets {
				v, gerr := secrets.Get(keyringGroup, key)
				if gerr == nil {
					val, err = resolveStoredSecretForList(v, onePassword)
					if err != nil {
						return nil, err
					}
				}
			}
			rows = append(rows, ListRow{Key: key, Value: val, Group: group})
		} else {
			rows = append(rows, ListRow{Key: key, Value: values[key], Group: group})
		}
	}
	for _, key := range secretKeys {
		if seen[key] {
			continue
		}
		val := "***"
		if showSecrets {
			v, gerr := secrets.Get(keyringGroup, key)
			if gerr == nil {
				val, err = resolveStoredSecretForList(v, onePassword)
				if err != nil {
					return nil, err
				}
			}
		}
		rows = append(rows, ListRow{Key: key, Value: val, Group: group})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Key < rows[j].Key })
	return rows, nil
}

func resolveStoredSecretForList(stored string, onePassword ports.OnePasswordEnv) (string, error) {
	if !strings.HasPrefix(stored, "op://") {
		return stored, nil
	}
	if onePassword == nil {
		return "", fmt.Errorf(
			"referência 1Password (op://) sem integração disponível (use sessão 1Password e cliente op)",
		)
	}
	return onePassword.ReadOPReference(stored)
}

func isSecretKey(secretKeys []string, key string) bool {
	for _, sk := range secretKeys {
		if sk == key {
			return true
		}
	}
	return false
}

func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
