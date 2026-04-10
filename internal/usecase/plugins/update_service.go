package plugins

import (
	"context"
	"fmt"
	"path/filepath"

	"mb/internal/ports"
)

// UpdateService handles updating of installed plugins.
type UpdateService struct {
	rt      PluginRuntime
	store   ports.PluginCLIStore
	scanner ports.PluginScanner
	shell   ports.ShellHelperInstaller
	git     ports.GitOperations
	fsys    ports.Filesystem
	syncer  *SyncService
}

// NewUpdateService creates a new UpdateService.
func NewUpdateService(
	rt PluginRuntime,
	store ports.PluginCLIStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	git ports.GitOperations,
	fsys ports.Filesystem,
	syncer *SyncService,
) *UpdateService {
	return &UpdateService{
		rt:      rt,
		store:   store,
		scanner: scanner,
		shell:   shell,
		git:     git,
		fsys:    fsys,
		syncer:  syncer,
	}
}

// UpdateRequest holds the parameters for updating plugins.
type UpdateRequest struct {
	// Package is a specific package name, or empty for all.
	Package string
}

// Update updates all remote Git plugins or a single one.
func (s *UpdateService) Update(ctx context.Context, req UpdateRequest, log Logger) error {
	_ = log.Info(ctx, "Atualizando plugins...")

	if req.Package != "" {
		return s.updateOne(ctx, req.Package, true, log)
	}

	sources, err := s.store.ListPluginSources()
	if err != nil {
		return err
	}
	for _, src := range sources {
		if src.GitURL == "" || src.LocalPath != "" {
			continue
		}
		if err := s.updateOne(ctx, src.InstallDir, true, log); err != nil {
			_ = log.Error(ctx, "%s: %v", src.InstallDir, err)
		}
	}

	opts := SyncOptions{EmitSuccess: false, NoRemove: false}
	_, err = s.syncer.Sync(ctx, opts, log)
	return err
}

func (s *UpdateService) updateOne(
	ctx context.Context,
	pkg string,
	logAlreadyLatest bool,
	log Logger,
) error {
	src, err := s.store.GetPluginSource(pkg)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("pacote %q não encontrado no registry", pkg)
	}
	if src.LocalPath != "" {
		return fmt.Errorf("pacote %q é local; não é possível atualizar", pkg)
	}
	if src.GitURL == "" {
		return fmt.Errorf(
			"pacote %q foi instalado manualmente (sem URL Git); não é possível atualizar",
			pkg,
		)
	}

	dir := filepath.Join(s.rt.PluginsDir, pkg)
	if !s.git.IsGitRepo(dir) {
		return fmt.Errorf("%s não é um repositório git", dir)
	}

	if src.RefType == "tag" {
		if err := s.git.FetchTags(ctx, dir); err != nil {
			return err
		}
		tags, err := s.git.ListLocalTags(dir)
		if err != nil {
			return err
		}
		var newerTag string
		for _, t := range tags {
			if _, isNewer := s.git.NewerTag(src.Ref, t); isNewer {
				newerTag = t
				break
			}
		}
		if newerTag == "" {
			if logAlreadyLatest && log != nil {
				_ = log.Info(ctx, "%s: já está na versão mais recente (%s)", pkg, src.Ref)
			}
			return nil
		}
		if err := s.git.CheckoutTag(ctx, dir, newerTag); err != nil {
			return err
		}
		version, _ := s.git.GetVersion(dir)
		src.Ref = newerTag
		src.Version = version
		if err := s.store.UpsertPluginSource(*src); err != nil {
			return err
		}
		_ = log.Info(ctx, "%s atualizado para %s", pkg, version)
		return nil
	}

	if err := s.git.FetchAndPull(ctx, dir, src.Ref); err != nil {
		return err
	}
	version, err := s.git.GetVersion(dir)
	if err != nil {
		return err
	}
	src.Version = version
	if err := s.store.UpsertPluginSource(*src); err != nil {
		return err
	}
	_ = log.Info(ctx, "%s atualizado para %s", pkg, version)
	return nil
}
