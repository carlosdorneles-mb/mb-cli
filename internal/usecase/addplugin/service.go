// Package addplugin provides the use case for installing a plugin from a
// remote Git URL or a local directory tree.
//
// It orchestrates validation, cloning, persistence, and sync without any
// dependency on Cobra or CLI concerns.
package addplugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
)

// Runtime holds configuration shared by plugin use cases.
type Runtime struct {
	ConfigDir  string
	PluginsDir string
}

// SyncOptions configures the post-add sync behaviour.
type SyncOptions struct {
	EmitSuccess bool
	NoRemove    bool
	PostSync    func(context.Context) error
}

// Logger is the minimal logging surface used by this use case.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any) error
	Warn(ctx context.Context, msg string, args ...any) error
	Debug(ctx context.Context, msg string, args ...any) error
	Error(ctx context.Context, msg string, args ...any) error
}

// SyncReport summarizes plugin command changes detected during sync.
type SyncReport struct {
	Added, Updated, Removed int
	AnyChange               bool
}

// Request holds the parameters for adding a plugin.
type Request struct {
	// Source is either a Git URL ("https://...", "git@...") or a local path (".", "/abs/path").
	Source string
	// Package overrides the auto-detected package name.
	Package string
	// Tag pins a specific Git tag to install (remote only).
	Tag string
	// NoRemove keeps SQLite rows for commands that disappeared.
	NoRemove bool
}

// Service orchestrates plugin addition.
// All dependencies are interfaces, making it trivially testable with fakes.
type Service struct {
	rt       Runtime
	store    ports.PluginCacheStore
	scanner  ports.PluginScanner
	fsys     ports.Filesystem
	git      ports.GitOperations
	shell    ports.ShellHelperInstaller
	layout   ports.PluginLayoutValidator
	syncer   *Syncer
}

// New creates a new AddPlugin service.
func New(rt Runtime, store ports.PluginCacheStore, scanner ports.PluginScanner, fsys ports.Filesystem, git ports.GitOperations, shell ports.ShellHelperInstaller, layout ports.PluginLayoutValidator, syncer *Syncer) *Service {
	return &Service{
		rt:     rt,
		store:  store,
		scanner: scanner,
		fsys:   fsys,
		git:    git,
		shell:  shell,
		layout: layout,
		syncer: syncer,
	}
}

// Add installs um plugin a partir de URL Git ou path local.
// The logger is passed per-call so each Cobra command can bind its own writer.
func (s *Service) Add(ctx context.Context, req Request, log Logger) error {
	source := strings.TrimSpace(req.Source)
	if source == "" {
		return fmt.Errorf("informe a URL do repositório, um path ou . para o diretório atual")
	}

	syncOpts := SyncOptions{EmitSuccess: false, NoRemove: req.NoRemove, PostSync: nil}

	_, _, err := s.git.ParseGitURL(source)
	if err == nil {
		return s.addRemote(ctx, source, req.Package, req.Tag, syncOpts, log)
	}
	return s.addLocal(ctx, source, req.Package, syncOpts, log)
}

func (s *Service) addRemote(ctx context.Context, gitURL, pkg, tag string, syncOpts SyncOptions, log Logger) error {
	repoName, normalizedURL, err := s.git.ParseGitURL(gitURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %w", err)
	}

	installDir := pkg
	if installDir == "" {
		installDir = repoName
	}

	destDir := filepath.Join(s.rt.PluginsDir, installDir)
	if s.dirExists(destDir) {
		if err := s.fsys.RemoveAll(destDir); err != nil {
			return fmt.Errorf("remover instalação anterior: %w", err)
		}
	}

	if err := s.fsys.MkdirAll(s.rt.PluginsDir, 0o755); err != nil {
		return err
	}

	opts := ports.GitCloneOpts{}
	if tag != "" {
		opts.BranchOrTag = tag
		opts.UseTag = true
	} else {
		latestTag, err := s.git.LatestTag(ctx, normalizedURL)
		if err != nil {
			return fmt.Errorf("listar tags: %w", err)
		}
		if latestTag != "" {
			opts.BranchOrTag = latestTag
			opts.UseTag = true
		}
	}

	if err := s.git.Clone(ctx, normalizedURL, destDir, opts); err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	version, err := s.git.GetVersion(destDir)
	if err != nil {
		_ = s.fsys.RemoveAll(destDir)
		return fmt.Errorf("obter versão: %w", err)
	}

	refType := "tag"
	ref := opts.BranchOrTag
	if !opts.UseTag {
		refType = "branch"
		ref, err = s.git.GetCurrentBranch(destDir)
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
	if err := s.store.UpsertPluginSource(ps); err != nil {
		_ = s.fsys.RemoveAll(destDir)
		return err
	}

	report, err := s.syncer.Run(ctx, s.toSyncRuntime(), s.store, s.scanner, s.shell, log, toSyncOptions(syncOpts))
	if err != nil {
		return err
	}
	if !report.AnyChange {
		_ = log.Info(ctx, "Pacote %q verificado; nenhum comando novo, atualizado ou removido.", installDir)
		return nil
	}
	_ = log.Info(ctx, "Pacote %q instalado em %s (versão %s)", installDir, destDir, version)
	return nil
}

func (s *Service) addLocal(ctx context.Context, pathArg, pkg string, syncOpts SyncOptions, log Logger) error {
	var absPath string
	var err error
	if pathArg == "." {
		absPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("obter diretório atual: %w", err)
		}
	} else {
		absPath, err = filepath.Abs(pathArg)
		if err != nil {
			return fmt.Errorf("caminho inválido: %w", err)
		}
	}

	info, err := s.fsys.Stat(absPath)
	if err != nil {
		if s.fsys.IsNotExist(err) {
			return fmt.Errorf("diretório não encontrado: %s", absPath)
		}
		return fmt.Errorf("acesso ao diretório: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("não é um diretório: %s", absPath)
	}

	rootManifest := filepath.Join(absPath, "manifest.yaml")
	if _, err := s.fsys.Stat(rootManifest); s.fsys.IsNotExist(err) {
		return s.addLocalCollection(ctx, absPath, pkg, syncOpts, log)
	} else if err != nil {
		return err
	}
	return s.addLocalSingle(ctx, absPath, pkg, syncOpts, log)
}

func (s *Service) addLocalCollection(ctx context.Context, absPath, pkg string, syncOpts SyncOptions, log Logger) error {
	entries, err := s.fsys.ReadDir(absPath)
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
		if _, err := s.fsys.Stat(filepath.Join(child, "manifest.yaml")); s.fsys.IsNotExist(err) {
			_ = log.Warn(ctx, "ignorando %q: sem manifest.yaml na raiz do subdiretório", e.Name())
			continue
		} else if err != nil {
			return err
		}
		if err := s.layout.ValidatePluginRoot(child); err != nil {
			_ = log.Warn(ctx, "ignorando %q: %v", e.Name(), err)
			continue
		}
		candidates = append(candidates, candidate{path: child, installDir: e.Name()})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].installDir < candidates[j].installDir
	})

	if len(candidates) == 0 {
		return fmt.Errorf("nenhum plugin encontrado: a raiz não tem manifest.yaml e nenhum subdiretório direto com manifest.yaml válido")
	}
	if pkg != "" && len(candidates) > 1 {
		return fmt.Errorf("--package não pode ser usado ao adicionar vários plugins de uma vez (%d encontrados)", len(candidates))
	}

	changed := 0
	for _, c := range candidates {
		installDir := c.installDir
		if len(candidates) == 1 && pkg != "" {
			installDir = pkg
		}
		existing, _ := s.store.GetPluginSource(installDir)
		if existing != nil {
			if err := s.store.UpsertPluginSource(plugin.PluginSource{InstallDir: installDir, LocalPath: c.path}); err != nil {
				return err
			}
			changed++
			continue
		}
		if s.dirExists(filepath.Join(s.rt.PluginsDir, installDir)) {
			installDir = s.uniqueInstallDir(installDir)
		}
		if err := s.store.UpsertPluginSource(plugin.PluginSource{InstallDir: installDir, LocalPath: c.path}); err != nil {
			return err
		}
		changed++
		_ = log.Info(ctx, "Pacote %q registrado localmente em %s", installDir, c.path)
	}
	if changed == 0 {
		return fmt.Errorf("nenhum pacote registado ou atualizado")
	}

	report, err := s.syncer.Run(ctx, s.toSyncRuntime(), s.store, s.scanner, s.shell, log, toSyncOptions(syncOpts))
	if err != nil {
		return err
	}
	if !report.AnyChange {
		_ = log.Info(ctx, "Pacotes verificados; nenhum comando novo, atualizado ou removido.")
	}
	return nil
}

func (s *Service) addLocalSingle(ctx context.Context, absPath, pkg string, syncOpts SyncOptions, log Logger) error {
	installDir := pkg
	if installDir == "" {
		installDir = filepath.Base(absPath)
	}
	existing, _ := s.store.GetPluginSource(installDir)
	if existing != nil {
		if err := s.store.UpsertPluginSource(plugin.PluginSource{InstallDir: installDir, LocalPath: absPath}); err != nil {
			return err
		}
		report, err := s.syncer.Run(ctx, s.toSyncRuntime(), s.store, s.scanner, s.shell, log, toSyncOptions(syncOpts))
		if err != nil {
			return err
		}
		if !report.AnyChange {
			_ = log.Info(ctx, "Pacote %q verificado; nenhum comando novo, atualizado ou removido.", installDir)
			return nil
		}
		_ = log.Info(ctx, "Pacote %q atualizado (path local: %s)", installDir, absPath)
		return nil
	}

	if s.dirExists(filepath.Join(s.rt.PluginsDir, installDir)) {
		installDir = s.uniqueInstallDir(installDir)
	}
	if err := s.store.UpsertPluginSource(plugin.PluginSource{InstallDir: installDir, LocalPath: absPath}); err != nil {
		return err
	}
	report, err := s.syncer.Run(ctx, s.toSyncRuntime(), s.store, s.scanner, s.shell, log, toSyncOptions(syncOpts))
	if err != nil {
		return err
	}
	if !report.AnyChange {
		_ = log.Info(ctx, "Pacote %q verificado; nenhum comando novo, atualizado ou removido.", installDir)
		return nil
	}
	_ = log.Info(ctx, "Pacote %q registrado localmente em %s", installDir, absPath)
	return nil
}

// --- helpers internos ---

func (s *Service) dirExists(p string) bool {
	info, err := s.fsys.Stat(p)
	return err == nil && info.IsDir()
}

func (s *Service) uniqueInstallDir(base string) string {
	for i := 2; ; i++ {
		installDir := fmt.Sprintf("%s-%d", base, i)
		destDir := filepath.Join(s.rt.PluginsDir, installDir)
		if !s.dirExists(destDir) {
			return installDir
		}
	}
}

func (s *Service) toSyncRuntime() Runtime {
	return Runtime{
		ConfigDir:  s.rt.ConfigDir,
		PluginsDir: s.rt.PluginsDir,
	}
}

func toSyncOptions(opts SyncOptions) SyncerOptions {
	return SyncerOptions{
		EmitSuccess: opts.EmitSuccess,
		NoRemove:    opts.NoRemove,
		PostSync:    opts.PostSync,
	}
}
