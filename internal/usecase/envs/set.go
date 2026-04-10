package envs

import (
	"fmt"
	"strings"

	"mb/internal/deps"
	"mb/internal/ports"
)

// Set persists key/value for the given group: plain file, keyring when asSecret,
// or 1Password (reference in keyring) when secretOP (--secret-op, sem --secret).
func Set(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
	setGroup, key, value string,
	asSecret, secretOP bool,
) error {
	if asSecret && secretOP {
		return fmt.Errorf("use apenas --secret ou --secret-op, não ambos")
	}
	path, err := TargetPath(paths, setGroup)
	if err != nil {
		return err
	}
	kg := KeyringGroup(setGroup)

	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return err
	}

	if asSecret || secretOP {
		var stored string
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
			stored = ref
		} else {
			stored = value
		}
		if err := secrets.Set(kg, key, stored); err != nil {
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
			raw, _ := secrets.Get(kg, key)
			if strings.HasPrefix(raw, "op://") && onePassword != nil {
				_ = onePassword.RemoveSecretField(kg, key)
			}
			_ = secrets.Delete(kg, key)
			_ = deps.RemoveSecretKey(path, key)
			break
		}
	}
	values[key] = value
	return deps.SaveDefaultEnvValues(path, values)
}
