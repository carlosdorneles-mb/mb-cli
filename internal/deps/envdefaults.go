package deps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
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

// BuildEnvFileValues loads env.defaults, overlays .env.<EnvGroup> when EnvGroup is set,
// then overlays --env-file. Used for plugin execution defaults.
func BuildEnvFileValues(rt *RuntimeConfig) (map[string]string, error) {
	merged := map[string]string{}

	defaultValues, err := LoadDefaultEnvValues(rt.DefaultEnvPath)
	if err != nil {
		return nil, err
	}
	for key, value := range defaultValues {
		merged[key] = value
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
