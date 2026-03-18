package plugins

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/deps"
	"mb/internal/commands/self"
	mbplugins "mb/internal/plugins"
	"mb/internal/ui"
)

func newPluginsAddCmd(deps deps.Dependencies) *cobra.Command {
	var name string
	var tag string

	cmd := &cobra.Command{
		Use:   "add <git-url|path|.>",
		Short: "Instala um plugin a partir de uma URL Git (remoto) ou de um diretório local (path ou .)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := strings.TrimSpace(args[0])
			// URL = remoto; path ou "." = local
			_, _, err := mbplugins.ParseGitURL(arg)
			if err == nil {
				return runAddRemote(cmd, deps, arg, name, tag)
			}
			return runAddLocal(cmd, deps, arg, name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Nome do plugin (diretório de instalação). Se não informado, usa o nome do repositório ou do diretório.")
	cmd.Flags().StringVar(&tag, "tag", "", "Instalar uma tag específica (apenas para plugin remoto).")
	return cmd
}

func runAddRemote(cmd *cobra.Command, deps deps.Dependencies, gitURL string, name string, tag string) error {
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

	if err := self.RunSync(deps, func(msg string) {}, cmd.ErrOrStderr()); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("plugin %q instalado em %s (versão %s)", installDir, destDir, version)))
	return nil
}

var errManifestFound = errors.New("manifest found")

func dirHasManifest(dir string) bool {
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "manifest.yaml" {
			return errManifestFound
		}
		return nil
	})
	return errors.Is(err, errManifestFound)
}

func runAddLocal(cmd *cobra.Command, deps deps.Dependencies, pathArg string, name string) error {
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
	if !dirHasManifest(absPath) {
		return fmt.Errorf("diretório não é um plugin válido: não contém manifest.yaml")
	}
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
	if err := deps.Store.UpsertPluginSource(cache.PluginSource{InstallDir: installDir, LocalPath: absPath}); err != nil {
		return err
	}
	if err := self.RunSync(deps, func(msg string) {}, cmd.ErrOrStderr()); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("O plugin %q foi registrado localmente em %s", installDir, absPath)))
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
