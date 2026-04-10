package plugins

import (
	"context"
	"fmt"
	"path/filepath"

	"mb/internal/ports"
)

// RemoveService handles removal of installed plugins.
type RemoveService struct {
	rt      PluginRuntime
	store   ports.PluginCLIStore
	scanner ports.PluginScanner
	shell   ports.ShellHelperInstaller
	fsys    ports.Filesystem
	syncer  *SyncService
}

// NewRemoveService creates a new RemoveService.
func NewRemoveService(
	rt PluginRuntime,
	store ports.PluginCLIStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	fsys ports.Filesystem,
	syncer *SyncService,
) *RemoveService {
	return &RemoveService{
		rt:      rt,
		store:   store,
		scanner: scanner,
		shell:   shell,
		fsys:    fsys,
		syncer:  syncer,
	}
}

// RemoveRequest holds the parameters for removing a plugin.
type RemoveRequest struct {
	Package string
}

// Remove uninstalls a plugin by package name and refreshes the cache.
func (s *RemoveService) Remove(ctx context.Context, req RemoveRequest, log Logger) error {
	src, err := s.store.GetPluginSource(req.Package)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("pacote %q não encontrado", req.Package)
	}

	if src.LocalPath == "" {
		destDir := filepath.Join(s.rt.PluginsDir, req.Package)
		if err := s.fsys.RemoveAll(destDir); err != nil {
			return fmt.Errorf("remover diretório: %w", err)
		}
	}
	if err := s.store.DeletePluginSource(req.Package); err != nil {
		return err
	}

	opts := SyncOptions{EmitSuccess: false, NoRemove: false}
	if _, err := s.syncer.Sync(ctx, opts, log); err != nil {
		return err
	}
	_ = log.Info(ctx, "Pacote %q removido", req.Package)
	return nil
}
