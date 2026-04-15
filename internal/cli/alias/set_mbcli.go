package alias

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

func runSetMbcliYAML(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	mbcliPath, name string,
	args []string,
	vaultChanged bool,
	vaultRaw string,
	yes bool,
) error {
	proj, err := deps.ParseMbcliAliases(mbcliPath)
	if err != nil {
		return err
	}
	var vaultInner string
	if vaultChanged {
		vaultInner, err = normalizeMbcliAliasVaultFlag(vaultRaw)
		if err != nil {
			return err
		}
	}
	slotVault := ""
	if vaultChanged {
		slotVault = vaultInner
	}
	sk := alib.StoreKey(slotVault, name)
	eProj, hadProj := proj[sk]

	if len(args) == 1 {
		if !vaultChanged {
			return errors.New(
				"indique --vault para atualizar só o vault, ou o comando após o nome " +
					"(ex.: mb alias set <nome> --mbcli-yaml -- <cmd>)",
			)
		}
		oldKey, eOne, n := alib.FindUniqueStoreKeyForDisplayName(proj, name)
		if n == 0 {
			return fmt.Errorf(
				"alias %q não existe em mbcli.yaml; defina com mb alias set %s --mbcli-yaml -- <cmd>",
				name, name,
			)
		}
		if n > 1 {
			return fmt.Errorf(
				"há vários aliases com o nome %q em mbcli.yaml; use mb alias set %s --mbcli-yaml --vault <vault-do-slot> -- <cmd>",
				name, name,
			)
		}
		eProj = eOne
		oldV := eProj.EnvVault
		prompt := fmt.Sprintf(
			"Deseja alterar o vault do alias %q (mbcli.yaml) para %s?",
			name,
			vaultDisplayForPrompt(true, vaultInner),
		)
		if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
			if errors.Is(err, ErrAliasOpCancelled) {
				_ = log.Info(ctx, "Operação cancelada.")
				return nil
			}
			return err
		}
		delete(proj, oldKey)
		eProj.EnvVault = vaultInner
		if err := alib.ValidateEntry(eProj); err != nil {
			return err
		}
		proj[alib.StoreKey(eProj.EnvVault, name)] = eProj
		if err := deps.WriteMbcliYAMLAliasSection(mbcliPath, proj); err != nil {
			return err
		}
		feedbackVaultOnly(ctx, log, name, oldV, vaultInner, true)
		_ = log.Info(ctx, "Salvo em %q.", mbcliPath)
		return nil
	}

	cmdArgv := append([]string(nil), args[1:]...)
	newE := alib.Entry{Command: cmdArgv}
	if vaultChanged {
		newE.EnvVault = vaultInner
	} else if hadProj {
		newE.EnvVault = eProj.EnvVault
	}
	if err := alib.ValidateEntry(newE); err != nil {
		return err
	}

	if hadProj {
		if err := confirmExistingAliasChange(ctx, cmd, yes, name, eProj, newE, true); err != nil {
			if errors.Is(err, ErrAliasOpCancelled) {
				_ = log.Info(ctx, "Operação cancelada.")
				return nil
			}
			return err
		}
	}

	if err := deps.UpsertMbcliYAMLAlias(mbcliPath, name, newE); err != nil {
		return err
	}
	if !hadProj {
		feedbackNewAlias(ctx, log, name, newE, true)
	} else {
		feedbackUpdatedAlias(ctx, log, name, eProj, newE, true)
	}
	_ = log.Info(ctx, "Salvo em %q.", mbcliPath)
	return nil
}
