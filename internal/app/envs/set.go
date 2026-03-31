package envs

import (
	"mb/internal/deps"
	"mb/internal/ports"
)

// Set persists key/value for the given group: plain file, or keyring when asSecret is true.
func Set(secrets ports.SecretStore, paths Paths, setGroup, key, value string, asSecret bool) error {
	path, err := TargetPath(paths, setGroup)
	if err != nil {
		return err
	}
	kg := KeyringGroup(setGroup)

	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return err
	}

	if asSecret {
		if err := secrets.Set(kg, key, value); err != nil {
			return err
		}
		if err := deps.AddSecretKey(path, key); err != nil {
			return err
		}
		delete(values, key)
		return deps.SaveDefaultEnvValues(path, values)
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
	values[key] = value
	return deps.SaveDefaultEnvValues(path, values)
}
