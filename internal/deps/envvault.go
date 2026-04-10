package deps

import "mb/internal/shared/envvault"

// ValidateEnvVault returns an error if name is not a safe vault identifier.
func ValidateEnvVault(name string) error {
	return envvault.Validate(name)
}

// VaultEnvFilePath returns the path to <configDir>/.env.<vault> for a validated vault name.
func VaultEnvFilePath(configDir, vault string) (string, error) {
	return envvault.FilePath(configDir, vault)
}
