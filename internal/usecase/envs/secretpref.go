package envs

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// MBEnvsSecretStoreKey is read (after explicit CLI flags) from the environment or plain target env files.
const MBEnvsSecretStoreKey = "MB_ENVS_SECRET_STORE"

// ResolveSetSecretFlags returns asSecret and secretOP from explicit flags, then getenv, then merged plain env files (paths in order; later overrides earlier).
func ResolveSetSecretFlags(
	explicitSecret, explicitSecretOP bool,
	pathsToScan ...string,
) (asSecret, secretOP bool, err error) {
	if explicitSecret && explicitSecretOP {
		return false, false, fmt.Errorf("use apenas --secret ou --secret-op, não ambos")
	}
	if explicitSecret {
		return true, false, nil
	}
	if explicitSecretOP {
		return false, true, nil
	}
	raw := strings.TrimSpace(os.Getenv(MBEnvsSecretStoreKey))
	if raw == "" {
		merged := map[string]string{}
		for _, p := range pathsToScan {
			m, _ := godotenv.Read(p)
			for k, v := range m {
				merged[k] = v
			}
		}
		raw = strings.TrimSpace(merged[MBEnvsSecretStoreKey])
	}
	switch strings.ToLower(raw) {
	case "":
		return false, false, nil
	case "plain", "file", "local":
		return false, false, nil
	case "keyring", "secret":
		return true, false, nil
	case "op", "1password", "onepassword":
		return false, true, nil
	default:
		return false, false, fmt.Errorf(
			"%s: valor inválido %q (use plain, keyring ou op)",
			MBEnvsSecretStoreKey,
			raw,
		)
	}
}
