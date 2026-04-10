package envs

import (
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/deps"
	"mb/internal/shared/envvault"
)

const opSecretsSuffix = ".opsecrets"

// VaultRow is one env file location shown by mb envs vaults.
type VaultRow struct {
	Vault    string `json:"vault"`
	Path     string `json:"path"`
	EnvCount int    `json:"env_count"`
}

// CollectVaultRows returns the default env file, logical project vaults from mbcli.yaml, and each
// per-vault .env.<nome> under config (excluding reserved name "project").
func CollectVaultRows(paths Paths) ([]VaultRow, error) {
	mbcliPath, err := deps.ResolveMbcliYAMLPath()
	if err != nil {
		return nil, err
	}
	defYaml, byNested, err := deps.ParseMbcliProjectEnvs(mbcliPath)
	if err != nil {
		return nil, err
	}

	out := make([]VaultRow, 0, 8+len(byNested))
	nDefault, err := CountListableEnvKeys(paths.DefaultEnvPath)
	if err != nil {
		return nil, err
	}
	out = append(out, VaultRow{Vault: "default", Path: paths.DefaultEnvPath, EnvCount: nDefault})

	if len(defYaml) > 0 {
		out = append(out, VaultRow{Vault: "project", Path: mbcliPath, EnvCount: len(defYaml)})
	}
	subNames := make([]string, 0, len(byNested))
	for k := range byNested {
		subNames = append(subNames, k)
	}
	sort.Strings(subNames)
	for _, k := range subNames {
		out = append(out, VaultRow{
			Vault:    "project/" + k,
			Path:     mbcliPath,
			EnvCount: len(byNested[k]),
		})
	}

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
		if v == "" || envvault.ValidateConfigurableVault(v) != nil {
			continue
		}
		nc, err := CountListableEnvKeys(path)
		if err != nil {
			return nil, err
		}
		rest = append(rest, VaultRow{Vault: v, Path: path, EnvCount: nc})
	}
	sort.Slice(rest, func(i, j int) bool {
		if rest[i].Vault != rest[j].Vault {
			return rest[i].Vault < rest[j].Vault
		}
		return rest[i].Path < rest[j].Path
	})
	out = append(out, rest...)
	return out, nil
}
