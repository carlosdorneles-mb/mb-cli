package envs

import (
	"github.com/spf13/cobra"

	appenvs "mb/internal/app/envs"
	"mb/internal/deps"
	"mb/internal/shared/system"
)

func newSetCmd(d deps.Dependencies) *cobra.Command {
	var setGroup string
	var secret, secretOP bool
	cmd := &cobra.Command{
		Use:     "set <KEY> <VALUE>",
		Aliases: []string{"s"},
		Short:   "Define ou atualiza uma variável padrão ou pra um grupo específico",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			key, value := args[0], args[1]
			if err := appenvs.Set(
				d.SecretStore,
				d.OnePassword,
				envPaths(d),
				setGroup,
				key,
				value,
				secret,
				secretOP,
			); err != nil {
				return err
			}

			if setGroup != "" {
				if secretOP {
					_ = log.Info(ctx, "variável %q guardada no 1Password (grupo %q)", key, setGroup)
				} else {
					_ = log.Info(ctx, "variável %q salva no grupo %q", key, setGroup)
				}
			} else {
				if secretOP {
					_ = log.Info(ctx, "variável %q guardada no 1Password (grupo padrão)", key)
				} else {
					_ = log.Info(ctx, "variável %q salva no grupo padrão", key)
				}
			}
			return nil
		},
	}
	cmd.Flags().
		StringVar(&setGroup, "group", "", "Grava a variável no grupo informado ao invés do grupo padrão")
	cmd.Flags().
		BoolVar(&secret, "secret", false, "Guarda o valor no keyring do sistema em vez do ficheiro env")
	cmd.Flags().
		BoolVar(&secretOP, "secret-op", false, "Guarda o valor no 1Password (CLI op); a referência op:// fica no keyring")
	cmd.MarkFlagsMutuallyExclusive("secret", "secret-op")
	cmd.GroupID = "commands"
	return cmd
}
