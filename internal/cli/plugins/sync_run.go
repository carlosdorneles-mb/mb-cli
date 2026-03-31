package plugins

import (
	"context"

	appplugins "mb/internal/app/plugins"
	"mb/internal/deps"
	"mb/internal/infra/shellhelpers"
	"mb/internal/shared/system"
)

// RunSync rescans plugin trees and refreshes SQLite (plugins, categories, help groups); see app/plugins.RunSync.
// Used by mb plugins sync and after plugins add/remove/update.
func RunSync(
	ctx context.Context,
	d deps.Dependencies,
	log *system.Logger,
	opts appplugins.SyncOptions,
) (appplugins.SyncReport, error) {
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
