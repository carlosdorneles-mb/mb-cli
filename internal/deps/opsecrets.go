package deps

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const opSecretsSuffix = ".opsecrets"

// LoadOPSecretRefs reads KEY=op://... from path+opSecretsSuffix. Missing file yields empty map.
func LoadOPSecretRefs(path string) (map[string]string, error) {
	fpath := path + opSecretsSuffix
	values, err := godotenv.Read(fpath)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(values))
	for k, v := range values {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	return out, nil
}

// SetOPSecretRef writes or updates one op:// reference for key in path.opsecrets.
func SetOPSecretRef(path, key, opRef string) error {
	refs, err := LoadOPSecretRefs(path)
	if err != nil {
		return err
	}
	refs[key] = opRef
	return saveOPSecretRefs(path, refs)
}

// RemoveOPSecretRef removes key from path.opsecrets; deletes the file when empty.
func RemoveOPSecretRef(path, key string) error {
	refs, err := LoadOPSecretRefs(path)
	if err != nil {
		return err
	}
	delete(refs, key)
	return saveOPSecretRefs(path, refs)
}

func saveOPSecretRefs(path string, refs map[string]string) error {
	fpath := path + opSecretsSuffix
	if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
		return err
	}
	if len(refs) == 0 {
		_ = os.Remove(fpath)
		return nil
	}
	return godotenv.Write(refs, fpath)
}
