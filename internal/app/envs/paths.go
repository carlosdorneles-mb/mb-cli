package envs

import "mb/internal/shared/envgroup"

// Paths selects default and per-group env files under the MB config layout.
type Paths struct {
	DefaultEnvPath string
	ConfigDir      string
}

// TargetPath resolves the env file path for an optional group flag (empty = default).
func TargetPath(paths Paths, group string) (string, error) {
	if group == "" {
		return paths.DefaultEnvPath, nil
	}
	if err := envgroup.Validate(group); err != nil {
		return "", err
	}
	return envgroup.FilePath(paths.ConfigDir, group)
}

// KeyringGroup maps CLI group "" to the keyring namespace "default".
func KeyringGroup(groupFlag string) string {
	if groupFlag == "" {
		return "default"
	}
	return groupFlag
}
