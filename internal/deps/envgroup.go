package deps

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const maxEnvGroupLen = 64

var envGroupPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidateEnvGroup returns an error if name is not a safe group identifier.
func ValidateEnvGroup(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nome do grupo não pode ser vazio")
	}
	if name == "." || name == ".." || strings.Contains(name, "..") {
		return errors.New("nome do grupo inválido")
	}
	if len(name) > maxEnvGroupLen {
		return fmt.Errorf("nome do grupo excede %d caracteres", maxEnvGroupLen)
	}
	if !envGroupPattern.MatchString(name) {
		return errors.New("nome do grupo inválido: use apenas letras, números, ., _ e -")
	}
	return nil
}

// GroupEnvFilePath returns the path to <configDir>/.env.<group> for a validated group name.
func GroupEnvFilePath(configDir, group string) (string, error) {
	if err := ValidateEnvGroup(group); err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".env."+group), nil
}
