package commands

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func loadDefaultEnvValues(path string) (map[string]string, error) {
	values, err := godotenv.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	return values, nil
}

func saveDefaultEnvValues(path string, values map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return godotenv.Write(values, path)
}
