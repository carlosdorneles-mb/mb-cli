package plugins

import (
	"context"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/system"
)

func newPluginsSyncCmd(deps deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Rescaneia plugins e reconstrói o cache",
		Long:  "Rescaneia o diretório de plugins e os paths locais registrados, atualiza o cache SQLite e garante os helpers de shell em ~/.config/mb/lib/shell.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())
			return RunSync(ctx, deps, log, true)
		},
	}
}
