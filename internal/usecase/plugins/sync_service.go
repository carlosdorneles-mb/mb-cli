package plugins

import (
	"context"

	"mb/internal/ports"
	"mb/internal/shared/system"
)

// Logger is the minimal logging surface used by plugin use cases.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any) error
	Warn(ctx context.Context, msg string, args ...any) error
	Debug(ctx context.Context, msg string, args ...any) error
	Error(ctx context.Context, msg string, args ...any) error
}

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
	// Adapt Logger interface to *system.Logger expected by RunSync.
	sysLog := toSystemLogger(log)
	return RunSync(ctx, s.rt, s.store, s.scanner, s.shell, sysLog, opts)
}

// toSystemLogger adapts a Logger interface to *system.Logger.
// Returns nil if log is nil or already a *system.Logger.
func toSystemLogger(log Logger) *system.Logger {
	if log == nil {
		return nil
	}
	if sl, ok := log.(*system.Logger); ok {
		return sl
	}
	// For non-system loggers, return nil (RunSync handles nil gracefully).
	return nil
}
