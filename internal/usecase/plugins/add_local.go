package plugins

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
	"mb/internal/shared/system"
)

// RunAddLocalPath registers a local plugin tree or collection and runs sync.
// absPath must be an absolute directory containing manifest.yaml or subdirectories with manifests.
func RunAddLocalPath(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	fsys ports.Filesystem,
	layout ports.PluginLayoutValidator,
	log *system.Logger,
	absPath string,
	pkg string,
	syncOpts SyncOptions,
) error {
	info, err := fsys.Stat(absPath)
	if err != nil {
		if fsys.IsNotExist(err) {
			return fmt.Errorf("diretório não encontrado: %s", absPath)
		}
		return fmt.Errorf("acesso ao diretório: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("não é um diretório: %s", absPath)
	}

	rootManifest := filepath.Join(absPath, "manifest.yaml")
	if _, err := fsys.Stat(rootManifest); fsys.IsNotExist(err) {
		return runAddLocalCollection(
			ctx,
			rt,
			store,
			scanner,
			shell,
			fsys,
			layout,
			log,
			absPath,
			pkg,
			syncOpts,
		)
	} else if err != nil {
		return err
	}
	return runAddLocalSingle(ctx, rt, store, scanner, shell, log, fsys, absPath, pkg, syncOpts)
}

func runAddLocalCollection(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	fsys ports.Filesystem,
	layout ports.PluginLayoutValidator,
	log *system.Logger,
	absPath string,
	pkg string,
	syncOpts SyncOptions,
) error {
	entries, err := fsys.ReadDir(absPath)
	if err != nil {
		return err
	}
	type candidate struct {
		path       string
		installDir string
	}
	var candidates []candidate
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		child := filepath.Join(absPath, e.Name())
		if _, err := fsys.Stat(filepath.Join(child, "manifest.yaml")); fsys.IsNotExist(err) {
			if log != nil {
				_ = log.Warn(
					ctx,
					"ignorando %q: sem manifest.yaml na raiz do subdiretório",
					e.Name(),
				)
			}
			continue
		} else if err != nil {
			return err
		}
		if err := layout.ValidatePluginRoot(child); err != nil {
			if log != nil {
				_ = log.Warn(ctx, "ignorando %q: %v", e.Name(), err)
			}
			continue
		}
		candidates = append(candidates, candidate{path: child, installDir: e.Name()})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].installDir < candidates[j].installDir
	})
	if len(candidates) == 0 {
		return fmt.Errorf(
			"nenhum plugin encontrado: a raiz não tem manifest.yaml e nenhum subdiretório direto com manifest.yaml válido",
		)
	}
	if pkg != "" && len(candidates) > 1 {
		return fmt.Errorf(
			"--package não pode ser usado ao adicionar vários plugins de uma vez (%d encontrados)",
			len(candidates),
		)
	}

	changed := 0
	for _, c := range candidates {
		installDir := c.installDir
		if len(candidates) == 1 && pkg != "" {
			installDir = pkg
		}
		existing, _ := store.GetPluginSource(installDir)
		if existing != nil {
			if err := store.UpsertPluginSource(
				plugin.PluginSource{InstallDir: installDir, LocalPath: c.path},
			); err != nil {
				return err
			}
			changed++
			continue
		}
		if dirExistsFS(fsys, filepath.Join(rt.PluginsDir, installDir)) {
			installDir = uniqueInstallDir(fsys, rt.PluginsDir, installDir)
		}
		if err := store.UpsertPluginSource(
			plugin.PluginSource{InstallDir: installDir, LocalPath: c.path},
		); err != nil {
			return err
		}
		changed++
		if log != nil {
			_ = log.Info(ctx, "Pacote %q registrado localmente em %s", installDir, c.path)
		}
	}
	if changed == 0 {
		return fmt.Errorf("nenhum pacote registado ou atualizado")
	}
	report, err := RunSync(ctx, rt, store, scanner, shell, log, syncOpts)
	if err != nil {
		return err
	}
	if !report.AnyChange && log != nil {
		_ = log.Info(ctx, "Pacotes verificados; nenhum comando novo, atualizado ou removido.")
	}
	return nil
}

func runAddLocalSingle(
	ctx context.Context,
	rt PluginRuntime,
	store ports.PluginCacheStore,
	scanner ports.PluginScanner,
	shell ports.ShellHelperInstaller,
	log *system.Logger,
	fsys ports.Filesystem,
	absPath string,
	pkg string,
	syncOpts SyncOptions,
) error {
	installDir := pkg
	if installDir == "" {
		installDir = filepath.Base(absPath)
	}
	existing, _ := store.GetPluginSource(installDir)
	if existing != nil {
		if err := store.UpsertPluginSource(
			plugin.PluginSource{InstallDir: installDir, LocalPath: absPath},
		); err != nil {
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
			_ = log.Info(ctx, "Pacote %q atualizado (path local: %s)", installDir, absPath)
		}
		return nil
	}
	if dirExistsFS(fsys, filepath.Join(rt.PluginsDir, installDir)) {
		installDir = uniqueInstallDir(fsys, rt.PluginsDir, installDir)
	}
	if err := store.UpsertPluginSource(
		plugin.PluginSource{InstallDir: installDir, LocalPath: absPath},
	); err != nil {
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
		_ = log.Info(ctx, "Pacote %q registrado localmente em %s", installDir, absPath)
	}
	return nil
}
