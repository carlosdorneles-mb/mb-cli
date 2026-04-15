package alias

import (
	"context"
	"slices"
	"strings"

	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

func feedbackNewAlias(
	ctx context.Context,
	log *system.Logger,
	name string,
	e alib.Entry,
	mbcliYAML bool,
) {
	if e.EnvVault == "" {
		if mbcliYAML {
			_ = log.Info(ctx, "Alias %q criado no vault project (raiz mbcli.yaml).", name)
		} else {
			_ = log.Info(ctx, "Alias %q criado sem vault associado.", name)
		}
		return
	}
	vDisp := e.EnvVault
	if mbcliYAML {
		vDisp = aliasMbcliLogicalVaultDisplay(e.EnvVault)
	}
	_ = log.Info(
		ctx,
		"Alias %q criado e associado ao vault %q. Ao executar `mb run %q` as variáveis do vault "+
			"serão injetadas no comando.",
		name, vDisp, name,
	)
}

func feedbackVaultOnly(
	ctx context.Context,
	log *system.Logger,
	name, oldV, newV string,
	mbcliLogical bool,
) {
	oldD, newD := oldV, newV
	if mbcliLogical {
		oldD = aliasMbcliLogicalVaultDisplay(oldV)
		newD = aliasMbcliLogicalVaultDisplay(newV)
	} else {
		oldD = configVaultLabel(oldV)
		newD = configVaultLabel(newV)
	}
	switch {
	case oldV == "" && newV != "":
		_ = log.Info(ctx, "Vault do alias %q definido como %q para uso com mb run.", name, newD)
	case oldV != "" && newV == "":
		_ = log.Info(
			ctx,
			"Vault do alias %q removido; mb run %q já não aplica um vault específico do alias.",
			name, name,
		)
	default:
		_ = log.Info(ctx, "Vault do alias %q alterado de %q para %q.", name, oldD, newD)
	}
}

func feedbackUpdatedAlias(
	ctx context.Context,
	log *system.Logger,
	name string,
	oldE, newE alib.Entry,
	mbcliYAML bool,
) {
	cmdChanged := !slices.Equal(oldE.Command, newE.Command)
	vaultChanged := oldE.EnvVault != newE.EnvVault

	switch {
	case cmdChanged && vaultChanged:
		_ = log.Info(ctx,
			"Alias %q atualizado: comando «%s» → «%s»; vault (mb run) %s → %s.",
			name,
			strings.Join(oldE.Command, " "),
			strings.Join(newE.Command, " "),
			vaultDisplayForPrompt(mbcliYAML, oldE.EnvVault),
			vaultDisplayForPrompt(mbcliYAML, newE.EnvVault),
		)
	case cmdChanged:
		_ = log.Info(
			ctx,
			"Alias %q atualizado: comando «%s» → «%s».",
			name,
			strings.Join(oldE.Command, " "),
			strings.Join(newE.Command, " "),
		)
	case vaultChanged:
		feedbackVaultOnly(ctx, log, name, oldE.EnvVault, newE.EnvVault, mbcliYAML)
	default:
		_ = log.Info(ctx, "Alias %q salvo (sem alterações efetivas).", name)
	}
}
