package plugins

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	mbfs "mb/internal/infra/fs"
	"mb/internal/infra/shellhelpers"
	"mb/internal/shared/system"
	appplugins "mb/internal/usecase/plugins"
)

func newPluginsRemoveCmd(d deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <package>",
		Aliases: []string{"rm", "r", "delete", "d", "del"},
		Short:   "Remove um plugin instalado",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			pkg := strings.TrimSpace(args[0])

			src, err := d.Store.GetPluginSource(pkg)
			if err != nil {
				return err
			}
			if src == nil {
				return fmt.Errorf("pacote %q não encontrado", pkg)
			}

			prompt := fmt.Sprintf("Remover o pacote %q?", pkg)
			confirmed, err := system.Confirm(ctx, prompt, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !confirmed {
				_ = log.Info(ctx, "Remoção cancelada")
				return nil
			}

			syncOpts := withCompletionPostSync(
				cmd,
				d,
				log,
				appplugins.SyncOptions{EmitSuccess: false},
			)
			return appplugins.RunRemovePackage(
				ctx,
				pluginRuntimeFromDeps(d),
				d.Store,
				d.Scanner,
				shellhelpers.Installer{},
				mbfs.OS{},
				log,
				pkg,
				syncOpts,
			)
		},
	}
}
