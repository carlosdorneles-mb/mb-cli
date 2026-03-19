package deps

import (
	"mb/internal/shared/envgroup"
)

// ValidateEnvGroup returns an error if name is not a safe group identifier.
func ValidateEnvGroup(name string) error {
	return envgroup.Validate(name)
}

// GroupEnvFilePath returns the path to <configDir>/.env.<group> for a validated group name.
func GroupEnvFilePath(configDir, group string) (string, error) {
	return envgroup.FilePath(configDir, group)
}
