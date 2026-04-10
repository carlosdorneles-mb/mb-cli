package envs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/deps"
	"mb/internal/ports"
	"mb/internal/shared/envvault"
)

const secretsSuffix = ".secrets"

// Storage values for ListRow.Storage (mb envs list table column).
const (
	StorageLocal     = "local"
	Storage1Password = "1password"
	StorageKeyring   = "keyring"
)

// ListRow is one variable shown by mb envs list.
type ListRow struct {
	Key, Value, Vault, Storage string
}

// CollectListRows returns rows for default and all valid vault env files, optionally revealing secrets.
func CollectListRows(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
	listVault string,
	showSecrets bool,
) ([]ListRow, error) {
	if listVault != "" {
		if err := envvault.Validate(listVault); err != nil {
			return nil, err
		}
		p, err := envvault.FilePath(paths.ConfigDir, listVault)
		if err != nil {
			return nil, err
		}
		return rowsForPath(secrets, onePassword, p, listVault, showSecrets)
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
		v := strings.TrimPrefix(base, ".env.")
		if v == "" || envvault.Validate(v) != nil {
			continue
		}
		if strings.HasSuffix(path, secretsSuffix) || strings.HasSuffix(path, opSecretsSuffix) {
			continue
		}
		r, err := rowsForPath(secrets, onePassword, path, v, showSecrets)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	sort.Slice(rows, func(i, j int) bool {
		vi, vj := rows[i].Vault, rows[j].Vault
		if vi != vj {
			if vi == "default" {
				return true
			}
			if vj == "default" {
				return false
			}
			return vi < vj
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
	path, vault string,
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
	opRefs, err := deps.LoadOPSecretRefs(path)
	if err != nil {
		return nil, err
	}
	keyringGroup := KeyringGroup(vault)

	keySet := make(map[string]bool)
	for k := range values {
		keySet[k] = true
	}
	for _, k := range secretKeys {
		keySet[k] = true
	}
	for k := range opRefs {
		keySet[k] = true
	}
	keys := sortedStringSet(keySet)

	var rows []ListRow
	for _, key := range keys {
		ref, inOP := opRefs[key]
		inOP = inOP && ref != "" && strings.HasPrefix(ref, "op://")
		inSK := isSecretKey(secretKeys, key)
		plainVal, inPlain := values[key]

		switch {
		case inOP:
			val := "***"
			if showSecrets {
				if onePassword == nil {
					return nil, fmt.Errorf(
						"referência 1Password (op://) sem integração disponível (use sessão 1Password e cliente op)",
					)
				}
				var rerr error
				val, rerr = onePassword.ReadOPReference(ref)
				if rerr != nil {
					return nil, rerr
				}
			}
			rows = append(rows, ListRow{
				Key: key, Value: val, Vault: vault, Storage: Storage1Password,
			})
		case inSK:
			stored, gerr := secrets.Get(keyringGroup, key)
			val := "***"
			if showSecrets && gerr == nil {
				var rerr error
				val, rerr = resolveStoredSecretForList(stored, onePassword)
				if rerr != nil {
					return nil, rerr
				}
			}
			rows = append(rows, ListRow{
				Key: key, Value: val, Vault: vault,
				Storage: secretStorageFromStored(stored, gerr),
			})
		case inPlain:
			rows = append(rows, ListRow{
				Key: key, Value: plainVal, Vault: vault, Storage: StorageLocal,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Key < rows[j].Key })
	return rows, nil
}

func sortedStringSet(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

func secretStorageFromStored(stored string, getErr error) string {
	if getErr != nil {
		return StorageKeyring
	}
	if strings.HasPrefix(stored, "op://") {
		return Storage1Password
	}
	return StorageKeyring
}

func isSecretKey(secretKeys []string, key string) bool {
	for _, sk := range secretKeys {
		if sk == key {
			return true
		}
	}
	return false
}
