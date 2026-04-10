package deps

import "mb/internal/shared/envvault"

// ValidateEnvVault returns an error if name is not allowed for ~/.config/mb/.env.<name> (incl. reservados).
func ValidateEnvVault(name string) error {
	return envvault.ValidateConfigurableVault(name)
}

// VaultEnvFilePath returns the path to <configDir>/.env.<vault> for a validated vault name.
func VaultEnvFilePath(configDir, vault string) (string, error) {
	return envvault.FilePath(configDir, vault)
}
