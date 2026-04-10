package plugins

import (
	"context"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	"mb/internal/usecase/plugins"
)

func newPluginsSyncCmd(svc *plugins.SyncService, d deps.Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Rescaneia plugins e reconstrói o cache",
		Long:  "Rescaneia o diretório de plugins e os paths locais registrados, atualiza o cache SQLite e garante os helpers de shell em ~/.config/mb/lib/shell.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			_, err := svc.Sync(ctx, plugins.SyncOptions{
				EmitSuccess: true,
			}, log)
			return err
		},
	}
	return cmd
}
