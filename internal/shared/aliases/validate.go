package aliases

import (
	"errors"
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

const maxNameLen = 64

// ValidateName checks alias names are safe for shell functions and YAML keys.
func ValidateName(name string) error {
	if name == "" {
		return errors.New("nome do alias não pode ser vazio")
	}
	if len(name) > maxNameLen {
		return fmt.Errorf("nome do alias excede %d caracteres", maxNameLen)
	}
	if !namePattern.MatchString(name) {
		return errors.New("nome do alias inválido: use apenas letras, números, _ e -")
	}
	return nil
}
