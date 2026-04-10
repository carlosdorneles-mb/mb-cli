package envs

import (
	"fmt"
	"strings"

	"mb/internal/deps"
	"mb/internal/ports"
)

// Set persists key/value for the given vault: plain file, keyring when asSecret,
// or 1Password with op:// reference stored in path.opsecrets when secretOP (--secret-op, sem --secret).
func Set(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
	setVault, key, value string,
	asSecret, secretOP bool,
) error {
	if asSecret && secretOP {
		return fmt.Errorf("use apenas --secret ou --secret-op, não ambos")
	}
	path, err := TargetPath(paths, setVault)
	if err != nil {
		return err
	}
	kg := KeyringGroup(setVault)

	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return err
	}

	if asSecret || secretOP {
		if secretOP {
			if onePassword == nil {
				return fmt.Errorf("integração 1Password indisponível")
			}
			if err := onePassword.EnsureAvailable(); err != nil {
				return err
			}
			ref, err := onePassword.PutSecret(kg, key, value)
			if err != nil {
				return err
			}
			if err := deps.SetOPSecretRef(path, key, ref); err != nil {
				return err
			}
			_ = deps.RemoveSecretKey(path, key)
			_ = secrets.Delete(kg, key)
		} else {
			if err := secrets.Set(kg, key, value); err != nil {
				return err
			}
			if err := deps.AddSecretKey(path, key); err != nil {
				return err
			}
			_ = deps.RemoveOPSecretRef(path, key)
		}
		delete(values, key)
		return deps.SaveDefaultEnvValues(path, values)
	}

	secretKeys, err := deps.LoadSecretKeys(path)
	if err != nil {
		return err
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
	opRefs, err := deps.LoadOPSecretRefs(path)
	if err != nil {
		return err
	}
	if _, ok := opRefs[key]; ok {
		if onePassword != nil {
			_ = onePassword.RemoveSecretField(kg, key)
		}
		_ = deps.RemoveOPSecretRef(path, key)
	}
	values[key] = value
	return deps.SaveDefaultEnvValues(path, values)
}
