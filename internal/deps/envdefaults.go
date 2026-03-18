package deps

import (
	"errors"
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
