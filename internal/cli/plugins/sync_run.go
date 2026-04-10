package plugins

import (
	"context"

	"github.com/spf13/cobra"

	"mb/internal/cli/completion"
	"mb/internal/cli/plugincmd"
	"mb/internal/deps"
	"mb/internal/infra/shellhelpers"
	"mb/internal/shared/system"
	appplugins "mb/internal/usecase/plugins"
)

// RunSync rescans plugin trees and refreshes SQLite (plugins, categories, help groups); see app/plugins.RunSync.
// Used by mb plugins sync and after plugins add/remove/update.
// cmd, quando não nil, permite atualizar o autocompletar do shell após sync bem-sucedido.
func RunSync(
	ctx context.Context,
	cmd *cobra.Command,
	d deps.Dependencies,
	log *system.Logger,
	opts appplugins.SyncOptions,
) (appplugins.SyncReport, error) {
	if cmd != nil {
		opts = withCompletionPostSync(cmd, d, log, opts)
	}
	return appplugins.RunSync(
		ctx,
		appplugins.PluginRuntime{
			ConfigDir:  d.Runtime.ConfigDir,
			PluginsDir: d.Runtime.PluginsDir,
			Quiet:      d.Runtime.Quiet,
			Verbose:    d.Runtime.Verbose,
		},
		d.Store,
		d.Scanner,
		shellhelpers.Installer{},
		log,
		opts,
	)
}

func withCompletionPostSync(
	cmd *cobra.Command,
	d deps.Dependencies,
	log *system.Logger,
	opts appplugins.SyncOptions,
) appplugins.SyncOptions {
	prev := opts.PostSync
	opts.PostSync = func(hookCtx context.Context) error {
		if prev != nil {
			if err := prev(hookCtx); err != nil {
				return err
			}
		}
		plugincmd.Reattach(cmd.Root(), d)
		return completion.TryRefreshInstalled(hookCtx, cmd.Root(), log, d.Runtime.Quiet)
	}
	return opts
}
