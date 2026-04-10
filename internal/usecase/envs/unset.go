package envs

import (
	"os"
	"strings"

	"mb/internal/deps"
	"mb/internal/ports"
)

// Unset removes a key from the env file and secret metadata for the given vault.
// It returns removed=false when the key was not defined for that vault.
func Unset(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
	vaultFlag, key string,
) (removed bool, err error) {
	path, err := TargetPath(paths, vaultFlag)
	if err != nil {
		return false, err
	}
	kg := KeyringGroup(vaultFlag)

	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return false, err
	}

	secretKeys, err := deps.LoadSecretKeys(path)
	if err != nil {
		return false, err
	}
	opRefs, err := deps.LoadOPSecretRefs(path)
	if err != nil {
		return false, err
	}

	_, inPlain := values[key]
	inSecrets := isSecretKey(secretKeys, key)
	_, inOP := opRefs[key]
	if !inPlain && !inSecrets && !inOP {
		return false, nil
	}

	if inOP {
		if onePassword != nil {
			_ = onePassword.RemoveSecretField(kg, key)
		}
		_ = deps.RemoveOPSecretRef(path, key)
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

	secretKeysAfter, err := deps.LoadSecretKeys(path)
	if err != nil {
		return false, err
	}
	opAfter, err := deps.LoadOPSecretRefs(path)
	if err != nil {
		return false, err
	}
	// Vault explícito: sem variáveis nem segredos, apagar .env.<vault> e sidecars.
	if vaultFlag != "" && len(values) == 0 && len(secretKeysAfter) == 0 && len(opAfter) == 0 {
		_ = os.Remove(path + ".secrets")
		_ = os.Remove(path + opSecretsSuffix)
		_ = os.Remove(path)
		return true, nil
	}

	if err := deps.SaveDefaultEnvValues(path, values); err != nil {
		return false, err
	}
	return true, nil
}
