package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/deps"
	mbplugins "mb/internal/plugins"
	"mb/internal/shared/system"
)

func newPluginsAddCmd(deps deps.Dependencies) *cobra.Command {
	var name string
	var tag string

	cmd := &cobra.Command{
		Use:     "add <git-url|path|.>",
		Aliases: []string{"install", "i", "a"},
		Short:   "Instala um plugin a partir de uma URL Git ou de um diretório local",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := strings.TrimSpace(args[0])
			// URL = remoto; path ou "." = local
			_, _, err := mbplugins.ParseGitURL(arg)
			if err == nil {
				return runAddRemote(cmd, deps, arg, name, tag)
			}
			log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())
			return runAddLocal(cmd, deps, log, arg, name)
		},
	}

	cmd.Flags().
		StringVar(&name, "name", "", "Nome do plugin (diretório de instalação). Se não informado, usa o nome do repositório ou do diretório.")
	cmd.Flags().
		StringVar(&tag, "tag", "", "Instalar uma tag específica (apenas para plugin remoto).")
	return cmd
}

func runAddRemote(
	cmd *cobra.Command,
	deps deps.Dependencies,
	gitURL string,
	name string,
	tag string,
) error {
	ctx := cmd.Context()

	repoName, normalizedURL, err := mbplugins.ParseGitURL(gitURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %w", err)
	}

	installDir := name
	if installDir == "" {
		installDir = repoName
	}

	destDir := filepath.Join(deps.Runtime.PluginsDir, installDir)
	if dirExists(destDir) {
		destDir, installDir = uniqueInstallDir(deps.Runtime.PluginsDir, installDir)
	}

	if err := os.MkdirAll(deps.Runtime.PluginsDir, 0o755); err != nil {
		return err
	}

	opts := mbplugins.CloneOpts{}
	if tag != "" {
		opts.BranchOrTag = tag
		opts.UseTag = true
	} else {
		latestTag, err := mbplugins.LatestTag(ctx, normalizedURL)
		if err != nil {
			return fmt.Errorf("listar tags: %w", err)
		}
		if latestTag != "" {
			opts.BranchOrTag = latestTag
			opts.UseTag = true
		}
	}

	if err := mbplugins.Clone(ctx, normalizedURL, destDir, opts); err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	version, err := mbplugins.GetVersion(destDir)
	if err != nil {
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("obter versão: %w", err)
	}

	refType := "tag"
	ref := opts.BranchOrTag
	if !opts.UseTag {
		refType = "branch"
		ref, err = mbplugins.GetCurrentBranch(destDir)
		if err != nil {
			ref = "main"
		}
	}
	if ref == "" {
		ref = version
	}

	ps := cache.PluginSource{
		InstallDir: installDir,
		GitURL:     gitURL,
		RefType:    refType,
		Ref:        ref,
		Version:    version,
	}
	if err := deps.Store.UpsertPluginSource(ps); err != nil {
		_ = os.RemoveAll(destDir)
		return err
	}

	log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())
	if err := RunSync(ctx, deps, log, false); err != nil {
		return err
	}
	_ = log.Info(ctx, "plugin %q instalado em %s (versão %s)", installDir, destDir, version)
	return nil
}

func runAddLocal(
	cmd *cobra.Command,
	deps deps.Dependencies,
	log *system.Logger,
	pathArg string,
	name string,
) error {
	if pathArg == "" {
		return fmt.Errorf("informe a URL do repositório, um path ou . para o diretório atual")
	}
	var absPath string
	if pathArg == "." {
		var err error
		absPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("obter diretório atual: %w", err)
		}
	} else {
		var err error
		absPath, err = filepath.Abs(pathArg)
		if err != nil {
			return fmt.Errorf("caminho inválido: %w", err)
		}
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("diretório não encontrado: %s", absPath)
		}
		return fmt.Errorf("acesso ao diretório: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("não é um diretório: %s", absPath)
	}

	rootManifest := filepath.Join(absPath, "manifest.yaml")
	if _, err := os.Stat(rootManifest); os.IsNotExist(err) {
		return runAddLocalCollection(cmd, deps, log, absPath, name)
	}
	return runAddLocalSingle(cmd, deps, log, absPath, name)
}

func runAddLocalCollection(
	cmd *cobra.Command,
	deps deps.Dependencies,
	log *system.Logger,
	absPath string,
	name string,
) error {
	ctx := cmd.Context()
	entries, err := os.ReadDir(absPath)
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
		if _, err := os.Stat(filepath.Join(child, "manifest.yaml")); os.IsNotExist(err) {
			_ = log.Warn(ctx, "ignorando %q: sem manifest.yaml na raiz do subdiretório", e.Name())
			continue
		}
		if err := mbplugins.ValidatePluginRoot(child); err != nil {
			_ = log.Warn(ctx, "ignorando %q: %v", e.Name(), err)
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
	if name != "" && len(candidates) > 1 {
		return fmt.Errorf(
			"--name não pode ser usado ao adicionar vários plugins de uma vez (%d encontrados)",
			len(candidates),
		)
	}

	added := 0
	for _, c := range candidates {
		installDir := c.installDir
		if len(candidates) == 1 && name != "" {
			installDir = name
		}
		existing, _ := deps.Store.GetPluginSource(installDir)
		if existing != nil {
			_ = log.Warn(ctx, "plugin %q já registrado, ignorando", installDir)
			continue
		}
		if dirExists(filepath.Join(deps.Runtime.PluginsDir, installDir)) {
			_, installDir = uniqueInstallDir(deps.Runtime.PluginsDir, installDir)
		}
		if err := deps.Store.UpsertPluginSource(
			cache.PluginSource{InstallDir: installDir, LocalPath: c.path},
		); err != nil {
			return err
		}
		added++
		_ = log.Info(ctx, "plugin %q registrado localmente em %s", installDir, c.path)
	}
	if added == 0 {
		return fmt.Errorf("nenhum plugin novo registrado (todos já existiam ou foram ignorados)")
	}
	if err := RunSync(ctx, deps, log, false); err != nil {
		return err
	}
	return nil
}

func runAddLocalSingle(
	cmd *cobra.Command,
	deps deps.Dependencies,
	log *system.Logger,
	absPath string,
	name string,
) error {
	ctx := cmd.Context()
	installDir := name
	if installDir == "" {
		installDir = filepath.Base(absPath)
	}
	existing, _ := deps.Store.GetPluginSource(installDir)
	if existing != nil {
		return fmt.Errorf("já existe um plugin com o nome %q", installDir)
	}
	if dirExists(filepath.Join(deps.Runtime.PluginsDir, installDir)) {
		_, installDir = uniqueInstallDir(deps.Runtime.PluginsDir, installDir)
	}
	if err := deps.Store.UpsertPluginSource(
		cache.PluginSource{InstallDir: installDir, LocalPath: absPath},
	); err != nil {
		return err
	}
	if err := RunSync(ctx, deps, log, false); err != nil {
		return err
	}
	_ = log.Info(ctx, "plugin %q registrado localmente em %s", installDir, absPath)
	return nil
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func uniqueInstallDir(pluginsDir, base string) (destDir string, installDir string) {
	for i := 2; ; i++ {
		installDir = fmt.Sprintf("%s-%d", base, i)
		destDir = filepath.Join(pluginsDir, installDir)
		if !dirExists(destDir) {
			return destDir, installDir
		}
	}
}
