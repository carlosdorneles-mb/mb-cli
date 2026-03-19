package update

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"mb/internal/commands/plugins"
	"mb/internal/commands/self"
	"mb/internal/deps"
	"mb/internal/system"
)

// NewUpdateCmd returns the root "mb update" command.
func NewUpdateCmd(d deps.Dependencies) *cobra.Command {
	var onlyPlugins, onlyCLI bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza plugins e o MB CLI",
		Long: `Atualiza, em sequência, os plugins instalados e o binário do MB CLI (conforme config).
Use --only-plugins para atualizar só os plugins; use --only-cli para atualizar só o binário.
Sem flags, executa as duas fases. Não use --only-plugins e --only-cli em simultâneo.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if onlyPlugins && onlyCLI {
				return errors.New("não use --only-plugins e --only-cli em simultâneo")
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(
				d.Runtime != nil && d.Runtime.Quiet,
				d.Runtime != nil && d.Runtime.Verbose,
				cmd.ErrOrStderr(),
			)
			runPlugins := !onlyCLI
			runCLI := !onlyPlugins

			if runPlugins {
				if d.Runtime != nil && !d.Runtime.Quiet {
					_ = log.Info(ctx, "Atualizando plugins...")
				}
				if err := plugins.RunUpdateAll(ctx, d, log); err != nil {
					return err
				}
			}
			if runCLI {
				if d.Runtime != nil && !d.Runtime.Quiet {
					_ = log.Info(ctx, "Atualizando MB CLI...")
				}
				return self.RunCLIUpdate(ctx, d, log)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&onlyPlugins, "only-plugins", false, "Atualiza apenas os plugins")
	cmd.Flags().BoolVar(&onlyCLI, "only-cli", false, "Atualiza apenas o MB CLI")
	cmd.GroupID = "commands"
	return cmd
}
