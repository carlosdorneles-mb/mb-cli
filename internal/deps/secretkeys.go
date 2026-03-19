package deps

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const secretsSuffix = ".secrets"

// LoadSecretKeys reads the list of key names stored in path+".secrets" (one key per line).
// Returns nil slice if the file does not exist.
func LoadSecretKeys(path string) ([]string, error) {
	fpath := path + secretsSuffix
	f, err := os.Open(fpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var keys []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		k := strings.TrimSpace(scanner.Text())
		if k != "" {
			keys = append(keys, k)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// AddSecretKey appends key to path+".secrets" if not already present.
func AddSecretKey(path, key string) error {
	if key == "" {
		return nil
	}
	keys, err := LoadSecretKeys(path)
	if err != nil {
		return err
	}
	for _, k := range keys {
		if k == key {
			return nil
		}
	}
	keys = append(keys, key)
	return saveSecretKeys(path, keys)
}

// RemoveSecretKey removes key from path+".secrets".
func RemoveSecretKey(path, key string) error {
	keys, err := LoadSecretKeys(path)
	if err != nil {
		return err
	}
	var newKeys []string
	for _, k := range keys {
		if k != key {
			newKeys = append(newKeys, k)
		}
	}
	return saveSecretKeys(path, newKeys)
}

func saveSecretKeys(path string, keys []string) error {
	fpath := path + secretsSuffix
	if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
		return err
	}
	if len(keys) == 0 {
		_ = os.Remove(fpath)
		return nil
	}
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, k := range keys {
		if _, err := f.WriteString(k + "\n"); err != nil {
			return err
		}
	}
	return nil
}
