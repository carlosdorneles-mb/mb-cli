package plugins

import (
	"context"
	"fmt"
	"path/filepath"

	"mb/internal/ports"
	"mb/internal/shared/system"
)

// RunUpdateAllGitPlugins pulls all remote Git sources and runs one sync at the end.
func RunUpdateAllGitPlugins(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	git ports.GitOperations,
	log *system.Logger,
	syncOpts SyncOptions,
) error {
	if log != nil {
		_ = log.Info(ctx, "Atualizando plugins...")
	}

	sources, err := store.ListPluginSources()
	if err != nil {
		return err
	}
	for _, src := range sources {
		if src.GitURL == "" || src.LocalPath != "" {
			continue
		}
		if err := updateOneRemotePackage(
			ctx,
			rt,
			store,
			git,
			log,
			src.InstallDir,
			true,
		); err != nil {
			if log != nil {
				_ = log.Error(ctx, "%s: %v", src.InstallDir, err)
			}
		}
	}
	_, err = RunSync(ctx, rt, store, scanner, shell, log, syncOpts)
	return err
}

// UpdateOneRemotePackage updates a single Git-backed package by install dir name (package id).
func UpdateOneRemotePackage(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	git ports.GitOperations,
	log *system.Logger,
	pkg string,
	logAlreadyLatest bool,
) error {
	return updateOneRemotePackage(ctx, rt, store, git, log, pkg, logAlreadyLatest)
}

func updateOneRemotePackage(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	git ports.GitOperations,
	log *system.Logger,
	pkg string,
	logAlreadyLatest bool,
) error {
	src, err := store.GetPluginSource(pkg)
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

	dir := filepath.Join(rt.PluginsDir, pkg)
	if !git.IsGitRepo(dir) {
		return fmt.Errorf("%s não é um repositório git", dir)
	}

	if src.RefType == "tag" {
		if err := git.FetchTags(ctx, dir); err != nil {
			return err
		}
		tags, err := git.ListLocalTags(dir)
		if err != nil {
			return err
		}
		var newerTag string
		for _, t := range tags {
			if _, isNewer := git.NewerTag(src.Ref, t); isNewer {
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
		if err := git.CheckoutTag(ctx, dir, newerTag); err != nil {
			return err
		}
		version, _ := git.GetVersion(dir)
		src.Ref = newerTag
		src.Version = version
		if err := store.UpsertPluginSource(*src); err != nil {
			return err
		}
		if log != nil {
			_ = log.Info(ctx, "%s atualizado para %s", pkg, version)
		}
		return nil
	}

	if err := git.FetchAndPull(ctx, dir, src.Ref); err != nil {
		return err
	}
	version, err := git.GetVersion(dir)
	if err != nil {
		return err
	}
	src.Version = version
	if err := store.UpsertPluginSource(*src); err != nil {
		return err
	}
	if log != nil {
		_ = log.Info(ctx, "%s atualizado para %s", pkg, version)
	}
	return nil
}
