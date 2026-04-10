package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	mbplugins "mb/internal/infra/plugins"
	"mb/internal/infra/shellhelpers"
	"mb/internal/shared/system"
	appplugins "mb/internal/usecase/plugins"
)

// RunUpdateAll updates all plugins that have a GitURL and no LocalPath, then runs sync.
func RunUpdateAll(
	ctx context.Context,
	cmd *cobra.Command,
	d deps.Dependencies,
	log *system.Logger,
) error {
	opts := appplugins.SyncOptions{EmitSuccess: false}
	if cmd != nil {
		opts = withCompletionPostSync(cmd, d, log, opts)
	}
	return appplugins.RunUpdateAllGitPlugins(
		ctx,
		pluginRuntimeFromDeps(d),
		d.Store,
		d.Scanner,
		shellhelpers.Installer{},
		mbplugins.GitService{},
		log,
		opts,
	)
}

func newPluginsUpdateCmd(d deps.Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:     "update <package>",
		Aliases: []string{"up", "u"},
		Short:   "Atualiza um plugin ou todos (--all)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

			if all {
				return RunUpdateAll(ctx, cmd, d, log)
			}

			if len(args) == 0 {
				return fmt.Errorf("informe o pacote ou use --all")
			}
			pkg := strings.TrimSpace(args[0])
			if err := appplugins.UpdateOneRemotePackage(
				ctx,
				pluginRuntimeFromDeps(d),
				d.Store,
				mbplugins.GitService{},
				log,
				pkg,
				true,
			); err != nil {
				return err
			}
			_, err := RunSync(ctx, cmd, d, log, appplugins.SyncOptions{EmitSuccess: true})
			return err
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Atualiza todos os plugins que tiverem nova versão")
	return cmd
}
