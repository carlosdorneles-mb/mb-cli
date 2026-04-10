package envvault

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const MaxNameLen = 64

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Validate returns an error if name is not a safe vault identifier.
func Validate(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nome do vault não pode ser vazio")
	}
	if name == "." || name == ".." || strings.Contains(name, "..") {
		return errors.New("nome do vault inválido")
	}
	if len(name) > MaxNameLen {
		return fmt.Errorf("nome do vault excede %d caracteres", MaxNameLen)
	}
	if !namePattern.MatchString(name) {
		return errors.New("nome do vault inválido: use apenas letras, números, ., _ e -")
	}
	return nil
}

// FilePath returns <configDir>/.env.<vault> for a validated vault name.
func FilePath(configDir, vault string) (string, error) {
	if err := Validate(vault); err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".env."+vault), nil
}
