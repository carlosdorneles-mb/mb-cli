package alias

import (
	"strings"

	"mb/internal/shared/envvault"
)

// aliasMbcliLogicalVaultDisplay maps the inner store vault for mbcli.yaml aliases
// to the same logical labels as mb envs list ("project", "project/<n>").
func aliasMbcliLogicalVaultDisplay(inner string) string {
	inner = strings.TrimSpace(inner)
	if inner == "" || inner == "default" {
		return "project"
	}
	return "project/" + inner
}

// configVaultLabel formats a config-dir vault name for prompts (empty -> "(nenhum)").
func configVaultLabel(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(nenhum)"
	}
	return v
}

// aliasListVaultDisplay is the vault column for mb alias list (table, JSON, fzf).
func aliasListVaultDisplay(source, inner string) string {
	if source == "project" {
		return aliasMbcliLogicalVaultDisplay(inner)
	}
	return configVaultLabel(inner)
}

// normalizeMbcliAliasVaultFlag converts --vault values for --mbcli-yaml to the inner
// vault key stored in aliases (empty = raiz envs no YAML; "staging" = submapa).
// Accepts logical names "project", "project/<n>", and plain "<n>" like mb envs set --mbcli-yaml.
func normalizeMbcliAliasVaultFlag(v string) (inner string, err error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}
	if v == "project" || v == "default" {
		return "", nil
	}
	if strings.HasPrefix(v, "project/") {
		suf := strings.TrimPrefix(v, "project/")
		if err := envvault.ValidateMbcliYAMLProjectNestedSuffix(suf); err != nil {
			return "", err
		}
		return suf, nil
	}
	if err := envvault.Validate(v); err != nil {
		return "", err
	}
	return v, nil
}
