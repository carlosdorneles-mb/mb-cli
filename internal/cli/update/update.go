package update

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cli/plugins"
	"mb/internal/deps"
	"mb/internal/infra/selfupdate"
	"mb/internal/shared/system"
	"mb/internal/shared/version"
)

// NewUpdateCmd returns the root "mb update" command.
func NewUpdateCmd(d deps.Dependencies) *cobra.Command {
	var onlyPlugins, onlyCLI, checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza o CLI, plugins e o sistema operacional (quando possível)",
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
				_ = log.Info(ctx, "Atualizando plugins...")
				if err := plugins.RunUpdateAll(ctx, d, log); err != nil {
					return err
				}
			}
			if runCLI {
				if checkOnly {
					suCfg := selfupdateFromAppConfig(d)
					local := strings.TrimSpace(version.Version)
					out, code, err := selfupdate.RunCheckOnly(ctx, suCfg, local)
					if out != "" {
						logInfoLines(ctx, log, out)
					}
					if err != nil {
						return err
					}
					if code == selfupdate.ExitCodeUpdateAvailable {
						os.Exit(selfupdate.ExitCodeUpdateAvailable)
					}
					return nil
				}
				_ = log.Info(ctx, "Atualizando MB CLI...")
				return RunCLIUpdate(ctx, d, log)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&onlyPlugins, "only-plugins", false, "Atualiza apenas os plugins")
	cmd.Flags().BoolVar(&onlyCLI, "only-cli", false, "Atualiza apenas o MB CLI")
	cmd.Flags().
		BoolVar(&checkOnly, "check-only", false, "Com --only-cli: só verifica se há atualização (sem baixar); saída 2 se houver")
	cmd.GroupID = "commands"
	return cmd
}
