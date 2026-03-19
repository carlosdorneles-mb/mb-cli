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
	var onlyPlugins, onlyCLI, onlySystem, checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza o CLI, plugins e o sistema operacional (quando possível)",
		Long: `Atualiza, em sequência, os plugins instalados, o binário do MB CLI (conforme config) e, sem nenhum --only-*, os pacotes do sistema (Homebrew/mas no macOS; apt/flatpak/snap no Linux quando disponíveis).

Sem flags, executa as três fases (plugins, CLI, sistema). Use --only-plugins, --only-cli e/ou --only-system para escolher quais fases correr; pode combinar várias (ex.: --only-plugins --only-cli).

--check-only só pode ser usado juntamente com --only-cli (verifica release do binário sem baixar).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if checkOnly && !onlyCLI {
				return errors.New("use --check-only apenas com --only-cli")
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

			runPlugins, runCLI, runSystem := resolveUpdatePhases(onlyPlugins, onlyCLI, onlySystem)

			if runPlugins {
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
				} else {
					if err := RunCLIUpdate(ctx, d, log); err != nil {
						return err
					}
				}
			}

			if runSystem {
				return RunSystemUpdate(ctx, log)
			}
			return nil
		},
	}
	cmd.Flags().
		BoolVar(&onlyPlugins, "only-plugins", false, "Inclui a fase de plugins (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&onlyCLI, "only-cli", false, "Inclui a fase do binário mb (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&onlySystem, "only-system", false, "Inclui a fase de pacotes do sistema (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&checkOnly, "check-only", false, "Só com --only-cli: verifica atualização do binário sem baixar; saída 2 se houver")
	return cmd
}

// resolveUpdatePhases returns which phases to run. If no --only-* flag is set, all three run.
func resolveUpdatePhases(onlyPlugins, onlyCLI, onlySystem bool) (plugins, cli, system bool) {
	if !onlyPlugins && !onlyCLI && !onlySystem {
		return true, true, true
	}
	return onlyPlugins, onlyCLI, onlySystem
}
