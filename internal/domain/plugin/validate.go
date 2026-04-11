package plugin

import (
	"regexp"
	"strconv"
	"strings"
)

// EnvFileEntry is the domain representation of a manifest env_files entry.
type EnvFileEntry struct {
	File  string
	Vault string
}

var helpGroupIDRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// ValidateManifestPure checks manifest fields without filesystem access.
// Returns warnings for issues that should be fixed but don't prevent loading.
func ValidateManifestPure(
	command, entrypoint string,
	hasFlags bool,
	envFiles []EnvFileEntry,
	groupID string,
	path string,
) []ValidationWarning {
	var warns []ValidationWarning

	// Command is required for executable leaves
	if (entrypoint != "" || hasFlags) && strings.TrimSpace(command) == "" {
		warns = append(warns, ValidationWarning{
			Path:    path,
			Message: "command é obrigatório quando há entrypoint ou flags",
		})
	}

	// Validate group_id format (nested leaves only)
	if groupID != "" {
		if !helpGroupIDRegex.MatchString(groupID) {
			warns = append(warns, ValidationWarning{
				Path:    path,
				Message: "group_id inválido: deve começar com letra e conter apenas letras, números, _ e -",
			})
		}
	}

	// Validate env_files entries
	for i, ef := range envFiles {
		if strings.TrimSpace(ef.File) == "" {
			warns = append(warns, ValidationWarning{
				Path:    path,
				Message: "env_files[" + strconv.Itoa(i) + "]: file não pode ser vazio",
			})
		}
	}

	return warns
}
