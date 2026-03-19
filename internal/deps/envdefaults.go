package deps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"mb/internal/keyring"
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
func mergeSecretKeysInto(merged map[string]string, path, keyringGroup string) {
	keys, err := LoadSecretKeys(path)
	if err != nil {
		return
	}
	for _, key := range keys {
		val, err := keyring.Get(keyringGroup, key)
		if err != nil {
			continue
		}
		merged[key] = val
	}
}

// BuildEnvFileValues loads env.defaults, overlays .env.<EnvGroup> when EnvGroup is set,
// then overlays --env-file. Secrets (keys in path.secrets) are resolved from the keyring.
// Used for plugin execution defaults.
func BuildEnvFileValues(rt *RuntimeConfig) (map[string]string, error) {
	merged := map[string]string{}

	defaultValues, err := LoadDefaultEnvValues(rt.DefaultEnvPath)
	if err != nil {
		return nil, err
	}
	for key, value := range defaultValues {
		merged[key] = value
	}
	mergeSecretKeysInto(merged, rt.DefaultEnvPath, "default")

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
		mergeSecretKeysInto(merged, groupPath, rt.EnvGroup)
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
