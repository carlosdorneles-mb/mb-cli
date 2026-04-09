package deps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"mb/internal/infra/keyring"
	"mb/internal/ports"
)

// LoadDefaultEnvValues reads KEY=VALUE pairs from the given path (e.g. env.defaults).
// Returns an empty map if the file does not exist.
func LoadDefaultEnvValues(path string) (map[string]string, error) {
	values, err := godotenv.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	return values, nil
}

// SaveDefaultEnvValues writes KEY=VALUE pairs to the given path, creating parent dirs if needed.
func SaveDefaultEnvValues(path string, values map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return godotenv.Write(values, path)
}

// mergeSecretKeysInto resolves keys listed in path.secrets from the keyring and adds them to merged.
// Values stored as op:// references are resolved via onePassword when non-nil.
func mergeSecretKeysInto(
	merged map[string]string,
	path, keyringGroup string,
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
) error {
	if secrets == nil {
		secrets = keyring.SystemKeyring{}
	}
	keys, err := LoadSecretKeys(path)
	if err != nil {
		return nil
	}
	for _, key := range keys {
		val, err := secrets.Get(keyringGroup, key)
		if err != nil {
			continue
		}
		resolved, rerr := resolveSecretValueForMerge(val, onePassword)
		if rerr != nil {
			return fmt.Errorf("variável %q (grupo %s): %w", key, keyringGroup, rerr)
		}
		merged[key] = resolved
	}
	return nil
}

func resolveSecretValueForMerge(val string, onePassword ports.OnePasswordEnv) (string, error) {
	if !strings.HasPrefix(val, "op://") {
		return val, nil
	}
	if onePassword == nil {
		return "", fmt.Errorf("referência 1Password (op://) sem integração disponível")
	}
	return onePassword.ReadOPReference(val)
}

// BuildEnvFileValues loads env.defaults, overlays .env.<EnvGroup> when EnvGroup is set,
// then overlays ./.env from the current working directory when that file exists,
// then overlays --env-file. Secrets (keys in path.secrets) are resolved from the keyring.
// Used for plugin execution and mb run.
// If secrets is nil, the OS keyring implementation is used.
// onePassword resolves op:// references; nil refuses op:// values with an error.
func BuildEnvFileValues(
	rt *RuntimeConfig,
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
) (map[string]string, error) {
	if secrets == nil {
		secrets = keyring.SystemKeyring{}
	}
	merged := map[string]string{}

	defaultValues, err := LoadDefaultEnvValues(rt.DefaultEnvPath)
	if err != nil {
		return nil, err
	}
	for key, value := range defaultValues {
		merged[key] = value
	}
	if err := mergeSecretKeysInto(
		merged,
		rt.DefaultEnvPath,
		"default",
		secrets,
		onePassword,
	); err != nil {
		return nil, err
	}

	if rt.EnvGroup != "" {
		if err := ValidateEnvGroup(rt.EnvGroup); err != nil {
			return nil, fmt.Errorf("--env-group: %w", err)
		}
		groupPath, err := GroupEnvFilePath(rt.ConfigDir, rt.EnvGroup)
		if err != nil {
			return nil, err
		}
		groupValues, err := LoadDefaultEnvValues(groupPath)
		if err != nil {
			return nil, err
		}
		for key, value := range groupValues {
			merged[key] = value
		}
		if err := mergeSecretKeysInto(
			merged,
			groupPath,
			rt.EnvGroup,
			secrets,
			onePassword,
		); err != nil {
			return nil, err
		}
	}

	if err := mergeCwdDotEnvIfPresent(merged); err != nil {
		return nil, fmt.Errorf(".env no diretório atual: %w", err)
	}

	if rt.EnvFilePath != "" {
		fileValues, readErr := godotenv.Read(rt.EnvFilePath)
		if readErr != nil {
			return nil, readErr
		}
		for key, value := range fileValues {
			merged[key] = value
		}
	}

	return merged, nil
}

// mergeCwdDotEnvIfPresent overlays ./.env from the current working directory when the file exists.
// It runs after env.defaults / --env-group and before --env-file.
func mergeCwdDotEnvIfPresent(merged map[string]string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("não foi possível obter o diretório atual: %w", err)
	}
	p := filepath.Join(wd, ".env")
	st, err := os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%q é um diretório, esperava-se um ficheiro", p)
	}
	values, err := godotenv.Read(p)
	if err != nil {
		return err
	}
	for k, v := range values {
		merged[k] = v
	}
	return nil
}
