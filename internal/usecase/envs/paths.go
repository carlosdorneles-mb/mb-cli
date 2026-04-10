package envs

import "mb/internal/shared/envvault"

// Paths selects default and per-vault env files under the MB config layout.
type Paths struct {
	DefaultEnvPath string
	ConfigDir      string
}

// TargetPath resolves the env file path for an optional vault flag (empty = default).
func TargetPath(paths Paths, vault string) (string, error) {
	if vault == "" {
		return paths.DefaultEnvPath, nil
	}
	if err := envvault.ValidateConfigurableVault(vault); err != nil {
		return "", err
	}
	return envvault.FilePath(paths.ConfigDir, vault)
}

// KeyringGroup maps CLI vault "" to the keyring / 1Password item namespace "default".
func KeyringGroup(vaultFlag string) string {
	if vaultFlag == "" {
		return "default"
	}
	return vaultFlag
}
