package envs

import (
	"strings"

	"mb/internal/deps"
	"mb/internal/ports"
)

// Unset removes a key from the env file and keyring metadata for the given group.
// It returns removed=false when the key was not defined for that group (neither in the
// env file nor in the group's secret key list).
func Unset(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
	groupFlag, key string,
) (removed bool, err error) {
	path, err := TargetPath(paths, groupFlag)
	if err != nil {
		return false, err
	}
	kg := KeyringGroup(groupFlag)

	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return false, err
	}

	secretKeys, err := deps.LoadSecretKeys(path)
	if err != nil {
		return false, err
	}
	if _, inFile := values[key]; !inFile && !isSecretKey(secretKeys, key) {
		return false, nil
	}

	for _, sk := range secretKeys {
		if sk != key {
			continue
		}
		raw, _ := secrets.Get(kg, key)
		if strings.HasPrefix(raw, "op://") && onePassword != nil {
			_ = onePassword.RemoveSecretField(kg, key)
		}
		_ = secrets.Delete(kg, key)
		_ = deps.RemoveSecretKey(path, key)
		break
	}
	delete(values, key)
	if err := deps.SaveDefaultEnvValues(path, values); err != nil {
		return false, err
	}
	return true, nil
}
