package plugins

import (
	"context"
	"fmt"
	"path/filepath"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
	"mb/internal/shared/system"
)

// RunAddRemote clones a Git URL into PluginsDir, registers the source, and runs sync.
func RunAddRemote(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	git ports.GitOperations,
	fsys ports.Filesystem,
	log *system.Logger,
	gitURL string,
	pkg string,
	tag string,
	syncOpts SyncOptions,
) error {
	repoName, normalizedURL, err := git.ParseGitURL(gitURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %w", err)
	}

	installDir := pkg
	if installDir == "" {
		installDir = repoName
	}

	destDir := filepath.Join(rt.PluginsDir, installDir)
	if dirExistsFS(fsys, destDir) {
		if err := fsys.RemoveAll(destDir); err != nil {
			return fmt.Errorf("remover instalação anterior: %w", err)
		}
	}

	if err := fsys.MkdirAll(rt.PluginsDir, 0o755); err != nil {
		return err
	}

	opts := ports.GitCloneOpts{}
	if tag != "" {
		opts.BranchOrTag = tag
		opts.UseTag = true
	} else {
		latestTag, err := git.LatestTag(ctx, normalizedURL)
		if err != nil {
			return fmt.Errorf("listar tags: %w", err)
		}
		if latestTag != "" {
			opts.BranchOrTag = latestTag
			opts.UseTag = true
		}
	}

	if err := git.Clone(ctx, normalizedURL, destDir, opts); err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	version, err := git.GetVersion(destDir)
	if err != nil {
		_ = fsys.RemoveAll(destDir)
		return fmt.Errorf("obter versão: %w", err)
	}

	refType := "tag"
	ref := opts.BranchOrTag
	if !opts.UseTag {
		refType = "branch"
		ref, err = git.GetCurrentBranch(destDir)
		if err != nil {
			ref = "main"
		}
	}
	if ref == "" {
		ref = version
	}

	ps := plugin.PluginSource{
		InstallDir: installDir,
		GitURL:     gitURL,
		RefType:    refType,
		Ref:        ref,
		Version:    version,
	}
	if err := store.UpsertPluginSource(ps); err != nil {
		_ = fsys.RemoveAll(destDir)
		return err
	}

	report, err := RunSync(ctx, rt, store, scanner, shell, log, syncOpts)
	if err != nil {
		return err
	}
	if !report.AnyChange {
		if log != nil {
			_ = log.Info(
				ctx,
				"Pacote %q verificado; nenhum comando novo, atualizado ou removido.",
				installDir,
			)
		}
		return nil
	}
	if log != nil {
		_ = log.Info(ctx, "Pacote %q instalado em %s (versão %s)", installDir, destDir, version)
	}
	return nil
}
