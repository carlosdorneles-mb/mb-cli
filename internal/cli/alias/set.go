package alias

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

func formatVaultLabel(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(nenhum)"
	}
	return v
}

func newSetCmd(d deps.Dependencies) *cobra.Command {
	var envVault string
	var yes bool

	cmd := &cobra.Command{
		Use:   "set <name> [--env-vault <vault>] [-- <command>...]",
		Short: "Define ou atualiza um alias",
		Long: `Associa um nome curto a um comando (com argumentos). Os dados ficam na configuração
do MB CLI e são refletidos nos scripts de shell gerados automaticamente.

Com --env-vault, apenas mb run <nome> usa esse vault de ambiente injetando as variáveis do vault ao comando executado.
Invocar o mesmo nome diretamente no shell, executa o comando com o ambiente normal do terminal sem a injeção
de variáveis de ambiente do vault.

Para alterar só o vault de um alias já existente, use --env-vault (e opcionalmente valor vazio
para desassociar o vault) com um único argumento: o nome do alias.

Alterar ou remover um alias existente pede confirmação no terminal; em CI ou scripts use --yes.`,
		Example: `
  # Define um alias para o comando docker compose up
  mb alias set dev -- docker compose up

  # Define um alias para o comando npm run build com o vault staging
  mb alias set dev --env-vault staging -- npm run build

  # Remove o vault associado ao alias
  mb alias set dev --env-vault ""

  # Atualiza o alias associado ao vault staging
  mb alias set dev --env-vault staging

  # Define um alias com valor literal para uma variável de ambiente
  mb alias set api --env-vault staging -- sh -c 'echo API $KEY'

  # Alias com comando em várias linhas (use \ no fim da linha antes de Enter no shell)
  mb alias set dev -- \
    docker compose \
    up -d
  
  # Sem prompt (automação); o comando também pode continuar em várias linhas
  mb alias set dev --yes -- \
    docker compose up -d`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

			name := args[0]
			if err := alib.ValidateName(name); err != nil {
				return err
			}

			cfgDir := d.Runtime.ConfigDir
			path := alib.FilePath(cfgDir)
			f, err := alib.Load(path)
			if err != nil {
				return err
			}

			envVaultFlag := cmd.Flags().Lookup("env-vault")
			envVaultChanged := envVaultFlag != nil && envVaultFlag.Changed

			if len(args) == 1 {
				if !envVaultChanged {
					return errors.New(
						"indique --env-vault para atualizar só o vault, ou o comando após o nome " +
							"(ex.: mb alias set <nome> -- <cmd>)",
					)
				}
				e, ok := f.Aliases[name]
				if !ok {
					hint := "defina o comando primeiro (ex.: mb alias set %s -- <cmd>)"
					return fmt.Errorf("alias %q ainda não existe; "+hint, name, name)
				}
				oldV := e.EnvVault
				prompt := fmt.Sprintf(
					"Deseja alterar o vault do alias %q para %s?",
					name,
					formatVaultLabel(envVault),
				)
				if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
					if errors.Is(err, ErrAliasOpCancelled) {
						_ = log.Info(ctx, "Operação cancelada.")
						return nil
					}
					return err
				}
				e.EnvVault = envVault
				if err := alib.ValidateEntry(e); err != nil {
					return err
				}
				f.Aliases[name] = e
				if err := saveAndRegenerate(cfgDir, f); err != nil {
					return err
				}
				feedbackVaultOnly(ctx, log, name, oldV, envVault)
				return nil
			}

			cmdArgv := append([]string(nil), args[1:]...)
			e, had := f.Aliases[name]
			newE := alib.Entry{Command: cmdArgv}
			if envVaultChanged {
				newE.EnvVault = envVault
			} else if had {
				newE.EnvVault = e.EnvVault
			}
			if err := alib.ValidateEntry(newE); err != nil {
				return err
			}

			if had {
				prompt := fmt.Sprintf(
					"O alias %q já existe e está associado ao vault %s.\n\nConfirmar alteração do comando e/ou vault?",
					name,
					formatVaultLabel(e.EnvVault),
				)
				if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
					if errors.Is(err, ErrAliasOpCancelled) {
						_ = log.Info(ctx, "Operação cancelada.")
						return nil
					}
					return err
				}
			}

			f.Aliases[name] = newE
			if err := saveAndRegenerate(cfgDir, f); err != nil {
				return err
			}

			if !had {
				feedbackNewAlias(ctx, log, name, newE)
				return nil
			}
			feedbackUpdatedAlias(ctx, log, name, e, newE)
			return nil
		},
	}

	cmd.Flags().StringVar(
		&envVault, "env-vault", "",
		"Vault de ambiente usado apenas com mb run; vazio remove a associação ao vault",
	)
	cmd.Flags().BoolVarP(
		&yes, "yes", "y", false,
		"Confirma alterações a aliases existentes sem prompt (CI / não interativo)",
	)
	return cmd
}

func feedbackNewAlias(ctx context.Context, log *system.Logger, name string, e alib.Entry) {
	if e.EnvVault == "" {
		_ = log.Info(ctx, "Alias %q criado sem vault associado.", name)
		return
	}
	_ = log.Info(
		ctx,
		"Alias %q criado e associado ao vault %q. Ao executar `mb run %q` as variáveis do vault "+
			"serão injetadas no comando.",
		name, e.EnvVault, name,
	)
}

func feedbackVaultOnly(ctx context.Context, log *system.Logger, name, oldV, newV string) {
	switch {
	case oldV == "" && newV != "":
		_ = log.Info(ctx, "Vault do alias %q definido como %q para uso com mb run.", name, newV)
	case oldV != "" && newV == "":
		_ = log.Info(
			ctx,
			"Vault do alias %q removido; mb run %q já não aplica um vault específico do alias.",
			name, name,
		)
	default:
		_ = log.Info(ctx, "Vault do alias %q alterado de %q para %q.", name, oldV, newV)
	}
}

func feedbackUpdatedAlias(
	ctx context.Context,
	log *system.Logger,
	name string,
	oldE, newE alib.Entry,
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
			formatVaultLabel(oldE.EnvVault),
			formatVaultLabel(newE.EnvVault),
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
		feedbackVaultOnly(ctx, log, name, oldE.EnvVault, newE.EnvVault)
	default:
		_ = log.Info(ctx, "Alias %q salvo (sem alterações efetivas).", name)
	}
}
