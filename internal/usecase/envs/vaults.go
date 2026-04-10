package envs

import (
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/shared/envvault"
)

const opSecretsSuffix = ".opsecrets"

// VaultRow is one env file location shown by mb envs vaults.
type VaultRow struct {
	Vault string `json:"vault"`
	Path  string `json:"path"`
}

// CollectVaultRows returns the default env file plus every per-vault .env.<nome> under config.
func CollectVaultRows(paths Paths) ([]VaultRow, error) {
	matches, err := filepath.Glob(filepath.Join(paths.ConfigDir, ".env.*"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var rest []VaultRow
	for _, path := range matches {
		if strings.HasSuffix(path, secretsSuffix) || strings.HasSuffix(path, opSecretsSuffix) {
			continue
		}
		base := filepath.Base(path)
		if !strings.HasPrefix(base, ".env.") {
			continue
		}
		v := strings.TrimPrefix(base, ".env.")
		if v == "" || envvault.Validate(v) != nil {
			continue
		}
		rest = append(rest, VaultRow{Vault: v, Path: path})
	}
	sort.Slice(rest, func(i, j int) bool {
		if rest[i].Vault != rest[j].Vault {
			return rest[i].Vault < rest[j].Vault
		}
		return rest[i].Path < rest[j].Path
	})
	out := make([]VaultRow, 0, 1+len(rest))
	out = append(out, VaultRow{Vault: "default", Path: paths.DefaultEnvPath})
	out = append(out, rest...)
	return out, nil
}
