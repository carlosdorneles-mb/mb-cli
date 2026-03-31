package plugins

import (
	"context"
	"fmt"
	"path/filepath"

	"mb/internal/ports"
	"mb/internal/shared/system"
)

// RunRemovePackage removes a registered package (disk + registry) and refreshes the cache.
// The caller must confirm with the user before invoking this.
func RunRemovePackage(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	fsys ports.Filesystem,
	log *system.Logger,
	pkg string,
	syncOpts SyncOptions,
) error {
	src, err := store.GetPluginSource(pkg)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("pacote %q não encontrado", pkg)
	}

	if src.LocalPath == "" {
		destDir := filepath.Join(rt.PluginsDir, pkg)
		if err := fsys.RemoveAll(destDir); err != nil {
			return fmt.Errorf("remover diretório: %w", err)
		}
	}
	if err := store.DeletePluginSource(pkg); err != nil {
		return err
	}
	if _, err := RunSync(
		ctx,
		rt,
		store,
		scanner,
		shell,
		log,
		syncOpts,
	); err != nil {
		return err
	}
	if log != nil {
		_ = log.Info(ctx, "Pacote %q removido", pkg)
	}
	return nil
}
