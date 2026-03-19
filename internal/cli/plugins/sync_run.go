package plugins

import (
	"context"

	appplugins "mb/internal/app/plugins"
	"mb/internal/deps"
	"mb/internal/shared/system"
)

// RunSync rescans the plugins dir and registered local paths, upserts plugins and categories, and updates the plugin_sources registry.
// Used by mb plugins sync and after plugins add/remove/update.
func RunSync(
	ctx context.Context,
	deps deps.Dependencies,
	log *system.Logger,
	opts appplugins.SyncOptions,
) (appplugins.SyncReport, error) {
	return appplugins.RunSync(ctx, deps, log, opts)
}
