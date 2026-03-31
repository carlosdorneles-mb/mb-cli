package envs

import (
	"mb/internal/deps"
	"mb/internal/ports"
)

// Unset removes a key from the env file and keyring metadata for the given group.
func Unset(secrets ports.SecretStore, paths Paths, groupFlag, key string) error {
	path, err := TargetPath(paths, groupFlag)
	if err != nil {
		return err
	}
	kg := KeyringGroup(groupFlag)

	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return err
	}

	secretKeys, err := deps.LoadSecretKeys(path)
	if err != nil {
		return err
	}
	for _, sk := range secretKeys {
		if sk == key {
			_ = secrets.Delete(kg, key)
			_ = deps.RemoveSecretKey(path, key)
			break
		}
	}
	delete(values, key)
	return deps.SaveDefaultEnvValues(path, values)
}
