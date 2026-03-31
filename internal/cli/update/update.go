// Package update implements the mb update command: phased updates for plugins, optional tools
// plugin (--update-all), CLI self-update, and OS package managers when available.
package update

import (
	"context"

	"github.com/spf13/cobra"

	appupdate "mb/internal/app/update"
	"mb/internal/cli/plugins"
	"mb/internal/deps"
	"mb/internal/shared/system"
)

// NewUpdateCmd builds the root "mb update" cobra command.
func NewUpdateCmd(d deps.Dependencies) *cobra.Command {
	var onlyPlugins, onlyCLI, onlySystem, onlyTools, checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza o CLI, plugins e o sistema operacional (quando possível)",
		Long: `Atualiza, em sequência, os plugins instalados, opcionalmente o agregador mb tools --update-all (se existir no cache), o binário do MB CLI (conforme config) e, sem nenhum --only-*, os pacotes do sistema (Homebrew/mas no macOS; apt/flatpak/snap no Linux quando disponíveis).

Sem flags, executa todas as fases habilitadas. Use --only-plugins, --only-tools, --only-cli e/ou --only-system para escolher quais fases correr; pode combinar várias (ex.: --only-plugins --only-cli).

--check-only só pode ser usado juntamente com --only-cli (verifica release do binário sem baixar).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(
				d.Runtime != nil && d.Runtime.Quiet,
				d.Runtime != nil && d.Runtime.Verbose,
				cmd.ErrOrStderr(),
			)
			return appupdate.Run(ctx, cmd, d, log, appupdate.Options{
				OnlyPlugins: onlyPlugins,
				OnlyCLI:     onlyCLI,
				OnlySystem:  onlySystem,
				OnlyTools:   onlyTools,
				CheckOnly:   checkOnly,
				RunAllGitPlugins: func(ctx context.Context) error {
					return plugins.RunUpdateAll(ctx, d, log)
				},
			})
		},
	}
	cmd.Flags().
		BoolVar(&onlyPlugins, "only-plugins", false, "Inclui a fase de plugins (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&onlyTools, "only-tools", false, "Inclui mb tools --update-all quando o plugin tools existir (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&onlyCLI, "only-cli", false, "Inclui a fase do binário mb (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&onlySystem, "only-system", false, "Inclui a fase de pacotes do sistema (combine com outros --only-*; sem nenhum, executa todas)")
	cmd.Flags().
		BoolVar(&checkOnly, "check-only", false, "Só com --only-cli: verifica atualização do binário sem baixar; saída 2 se houver")
	return cmd
}
