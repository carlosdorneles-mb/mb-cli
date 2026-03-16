package plugincmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/commands/self"
	mbplugins "mb/internal/plugins"
	"mb/internal/ui"
)

func newPluginsAddCmd(deps config.Dependencies) *cobra.Command {
	var name string
	var tag string

	cmd := &cobra.Command{
		Use:   "add <git-url>",
		Short: "Instala um plugin a partir de uma URL Git (GitHub, Bitbucket, GitLab)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			gitURL := strings.TrimSpace(args[0])

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
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Nome do plugin (diretório de instalação). Se não informado, usa o nome do repositório.")
	cmd.Flags().StringVar(&tag, "tag", "", "Instalar uma tag específica em vez da mais recente ou da branch main.")
	return cmd
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
