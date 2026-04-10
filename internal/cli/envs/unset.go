package envs

import (
	"github.com/spf13/cobra"

	appenvs "mb/internal/usecase/envs"
	"mb/internal/deps"
	"mb/internal/shared/system"
)

func newUnsetCmd(d deps.Dependencies) *cobra.Command {
	var unsetGroup string
	cmd := &cobra.Command{
		Use:     "unset <KEY>",
		Aliases: []string{"u"},
		Short:   "Remove uma variável do grupo padrão ou de um grupo específico",
		Long: `Remove a chave do ficheiro de ambiente do grupo escolhido.

Sem --group, o alvo é o grupo padrão.
Com --group <nome>, o alvo é o grupo específico.`,
		Example: `# Grupo padrão (env.defaults)
  mb envs unset API_URL

  # Grupo explícito (.env.staging)
  mb envs unset API_URL --group staging`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			key := args[0]
			removed, err := appenvs.Unset(
				d.SecretStore,
				d.OnePassword,
				envPaths(d),
				unsetGroup,
				key,
			)
			if err != nil {
				return err
			}
			if !removed {
				if unsetGroup != "" {
					_ = log.Info(ctx, "Não existe variável %q no grupo %q", key, unsetGroup)
				} else {
					_ = log.Info(ctx, "Não existe variável %q no grupo padrão", key)
				}
				return nil
			}
			if unsetGroup != "" {
				_ = log.Info(ctx, "Variável %q removida do grupo %q", key, unsetGroup)
			} else {
				_ = log.Info(ctx, "Variável %q removida do grupo padrão", key)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(
		&unsetGroup,
		"group",
		"",
		"Remove a variável do grupo informado ao invés do grupo padrão (ex.: --group staging)",
	)
	cmd.GroupID = "commands"
	return cmd
}
