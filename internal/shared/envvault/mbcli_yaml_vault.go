package envvault

import (
	"fmt"
	"strings"
)

// ValidateMbcliYAMLProjectNestedSuffix validates the segment after "project/" in user-facing
// mbcli.yaml vault flags (aliases --mbcli-yaml, mb envs list filters, etc.).
func ValidateMbcliYAMLProjectNestedSuffix(suf string) error {
	suf = strings.TrimSpace(suf)
	if suf == "" {
		return fmt.Errorf(
			`vault "project/" inválido: indique um nome após a barra (ex.: project/staging)`,
		)
	}
	if err := Validate(suf); err != nil {
		return fmt.Errorf("vault de projeto: %w", err)
	}
	return nil
}
