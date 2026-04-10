package envs

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	appenvs "mb/internal/usecase/envs"
)

func newUnsetCmd(d deps.Dependencies) *cobra.Command {
	var unsetVault string
	cmd := &cobra.Command{
		Use:     "unset <KEY> [<KEY>...]",
		Aliases: []string{"u"},
		Short:   "Remove variáveis do vault padrão ou de um vault específico",
		Long: `Remove as chaves do ficheiro de ambiente do vault escolhido.

Sem --vault, o alvo é o vault padrão.
Com --vault <nome>, o alvo é o ficheiro .env.<nome>.`,
		Example: `# Vault padrão (env.defaults)
  mb envs unset API_URL

  # Vault explícito (.env.staging)
  mb envs unset API_URL --vault staging

  # Várias chaves
  mb envs unset A B C`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			for _, key := range args {
				removed, err := appenvs.Unset(
					d.SecretStore,
					d.OnePassword,
					envPaths(d),
					unsetVault,
					key,
				)
				if err != nil {
					return err
				}
				if !removed {
					if unsetVault != "" {
						_ = log.Info(ctx, "Não existe variável %q no vault %q", key, unsetVault)
					} else {
						_ = log.Info(ctx, "Não existe variável %q no vault padrão", key)
					}
					continue
				}
				if unsetVault != "" {
					_ = log.Info(ctx, "Variável %q removida do vault %q", key, unsetVault)
				} else {
					_ = log.Info(ctx, "Variável %q removida do vault padrão", key)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(
		&unsetVault,
		"vault",
		"",
		"Remove do vault informado em vez do vault padrão (ex.: --vault staging)",
	)
	cmd.GroupID = "commands"
	return cmd
}
