package alias

import (
	"context"
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	alib "mb/internal/shared/aliases"
)

func vaultDisplayForPrompt(mbcliYAML bool, inner string) string {
	if mbcliYAML {
		return aliasMbcliLogicalVaultDisplay(inner)
	}
	return configVaultLabel(inner)
}

// confirmExistingAliasChange prompts only when command and/or vault actually change.
// For a true no-op (same Entry), it skips confirmation.
func confirmExistingAliasChange(
	ctx context.Context,
	cmd *cobra.Command,
	yes bool,
	name string,
	oldE, newE alib.Entry,
	mbcliYAML bool,
) error {
	cmdChanged := !slices.Equal(oldE.Command, newE.Command)
	vaultChanged := oldE.EnvVault != newE.EnvVault
	if !cmdChanged && !vaultChanged {
		return nil
	}

	var prompt string
	switch {
	case !cmdChanged && vaultChanged:
		if mbcliYAML {
			prompt = fmt.Sprintf(
				"Deseja alterar o vault do alias %q (mbcli.yaml) para %s?",
				name,
				vaultDisplayForPrompt(true, newE.EnvVault),
			)
		} else {
			prompt = fmt.Sprintf(
				"Deseja alterar o vault do alias %q para %s?",
				name,
				vaultDisplayForPrompt(false, newE.EnvVault),
			)
		}
	case cmdChanged && !vaultChanged:
		if mbcliYAML {
			prompt = fmt.Sprintf(
				"O alias %q já existe em mbcli.yaml (vault %s).\n\nConfirmar alteração do comando?",
				name,
				vaultDisplayForPrompt(true, oldE.EnvVault),
			)
		} else {
			prompt = fmt.Sprintf(
				"O alias %q já existe (vault %s).\n\nConfirmar alteração do comando?",
				name,
				vaultDisplayForPrompt(false, oldE.EnvVault),
			)
		}
	default:
		if mbcliYAML {
			prompt = fmt.Sprintf(
				"O alias %q já existe em mbcli.yaml e está associado ao vault %s.\n\nConfirmar alteração do comando e do vault?",
				name,
				vaultDisplayForPrompt(true, oldE.EnvVault),
			)
		} else {
			prompt = fmt.Sprintf(
				"O alias %q já existe e está associado ao vault %s.\n\nConfirmar alteração do comando e do vault?",
				name,
				vaultDisplayForPrompt(false, oldE.EnvVault),
			)
		}
	}
	return requireConfirmOrYes(ctx, cmd, yes, prompt)
}
