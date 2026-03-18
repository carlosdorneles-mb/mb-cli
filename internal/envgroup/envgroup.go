package envgroup

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const MaxNameLen = 64

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Validate returns an error if name is not a safe group identifier.
func Validate(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nome do grupo não pode ser vazio")
	}
	if name == "." || name == ".." || strings.Contains(name, "..") {
		return errors.New("nome do grupo inválido")
	}
	if len(name) > MaxNameLen {
		return fmt.Errorf("nome do grupo excede %d caracteres", MaxNameLen)
	}
	if !namePattern.MatchString(name) {
		return errors.New("nome do grupo inválido: use apenas letras, números, ., _ e -")
	}
	return nil
}

// FilePath returns <configDir>/.env.<group> for a validated group name.
func FilePath(configDir, group string) (string, error) {
	if err := Validate(group); err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".env."+group), nil
}
