package plugins

import (
	"context"

	"mb/internal/ports"
)

// Logger is the minimal logging surface used by plugin use cases.
type Logger = ports.Logger

// SyncService wraps RunSync so it can be injected as a dependency
// rather than called as a package-level function with raw deps.
type SyncService struct {
	rt      PluginRuntime
	store   ports.PluginCLIStore
	scanner ports.PluginScanner
	shell   ports.ShellHelperInstaller
}

// NewSyncService creates a new SyncService.
func NewSyncService(
	rt PluginRuntime,
	store ports.PluginCLIStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
) *SyncService {
	return &SyncService{
		rt:      rt,
		store:   store,
		scanner: scanner,
		shell:   shell,
	}
}

// Sync rescans PluginsDir and refreshes SQLite cache.
func (s *SyncService) Sync(ctx context.Context, opts SyncOptions, log Logger) (SyncReport, error) {
	return RunSync(ctx, s.rt, s.store, s.scanner, s.shell, log, opts)
}
