package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/plugins"
	"mb/internal/system"
	"mb/internal/ui"
)

func NewPluginsCmd(deps Dependencies) *cobra.Command {
	pluginsCmd := &cobra.Command{
		Use:     "plugins",
		Short:   "Gerencia plugins instalados (add, list, remove, update)",
		GroupID: "commands",
	}

	pluginsCmd.AddCommand(newPluginsAddCmd(deps))
	pluginsCmd.AddCommand(newPluginsListCmd(deps))
	pluginsCmd.AddCommand(newPluginsRemoveCmd(deps))
	pluginsCmd.AddCommand(newPluginsUpdateCmd(deps))
	return pluginsCmd
}

func newPluginsAddCmd(deps Dependencies) *cobra.Command {
	var name string
	var tag string

	cmd := &cobra.Command{
		Use:   "add <git-url>",
		Short: "Instala um plugin a partir de uma URL Git (GitHub, Bitbucket, GitLab)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			gitURL := strings.TrimSpace(args[0])

			repoName, normalizedURL, err := plugins.ParseGitURL(gitURL)
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

			opts := plugins.CloneOpts{}
			if tag != "" {
				opts.BranchOrTag = tag
				opts.UseTag = true
			} else {
				latestTag, err := plugins.LatestTag(ctx, normalizedURL)
				if err != nil {
					return fmt.Errorf("listar tags: %w", err)
				}
				if latestTag != "" {
					opts.BranchOrTag = latestTag
					opts.UseTag = true
				}
				// else clone default branch (no --branch)
			}

			if err := plugins.Clone(ctx, normalizedURL, destDir, opts); err != nil {
				return fmt.Errorf("clone: %w", err)
			}

			version, err := plugins.GetVersion(destDir)
			if err != nil {
				_ = os.RemoveAll(destDir)
				return fmt.Errorf("obter versão: %w", err)
			}

			refType := "tag"
			ref := opts.BranchOrTag
			if !opts.UseTag {
				refType = "branch"
				ref, err = plugins.GetCurrentBranch(destDir)
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

			if err := RunSync(deps, func(msg string) {}); err != nil {
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

func newPluginsListCmd(deps Dependencies) *cobra.Command {
	var checkUpdates bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista plugins instalados (name, command, description, version, url)",
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginList, err := deps.Store.ListPlugins()
			if err != nil {
				return err
			}
			sources, err := deps.Store.ListPluginSources()
			if err != nil {
				return err
			}
			sourceByDir := make(map[string]*cache.PluginSource)
			for i := range sources {
				sourceByDir[sources[i].InstallDir] = &sources[i]
			}

			sort.Slice(pluginList, func(i, j int) bool {
				return pluginList[i].CommandPath < pluginList[j].CommandPath
			})

			rows := make([][]string, 0, len(pluginList))
			for _, p := range pluginList {
				installDir := firstPathSegment(p.CommandPath)
				src := sourceByDir[installDir]
				name := installDir
				version := "-"
				url := "-"
				if src != nil {
					version = src.Version
					if src.GitURL != "" {
						url = src.GitURL
					}
				}

				updateAvail := ""
				if checkUpdates && src != nil && src.GitURL != "" {
					dir := filepath.Join(deps.Runtime.PluginsDir, installDir)
					if plugins.IsGitRepo(dir) {
						if src.RefType == "tag" {
							_ = plugins.FetchTags(cmd.Context(), dir)
							tags, _ := plugins.ListLocalTags(dir)
							for _, t := range tags {
								if _, newer := plugins.NewerTag(src.Ref, t); newer {
									updateAvail = "sim"
									break
								}
							}
						} else {
							// branch: could compare remote SHA; for now leave empty or "?"
							updateAvail = "-"
						}
					}
				}

				rows = append(rows, []string{name, p.CommandPath, p.Description, version, url, updateAvail})
			}

			headers := []string{"NAME", "COMMAND", "DESCRIPTION", "VERSION", "URL", "UPDATE"}
			if !checkUpdates {
				headers = []string{"NAME", "COMMAND", "DESCRIPTION", "VERSION", "URL"}
				for i := range rows {
					rows[i] = rows[i][:5]
				}
			}
			return system.Table(context.Background(), headers, rows)
		},
	}

	cmd.Flags().BoolVar(&checkUpdates, "check-updates", false, "Verifica se há atualização disponível para cada plugin")
	return cmd
}

func newPluginsRemoveCmd(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove um plugin instalado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			destDir := filepath.Join(deps.Runtime.PluginsDir, name)

			if !dirExists(destDir) {
				return fmt.Errorf("plugin %q não encontrado em %s", name, destDir)
			}

			confirmed, err := confirmRemove(cmd, name)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderInfo("remoção cancelada"))
				return nil
			}

			if err := os.RemoveAll(destDir); err != nil {
				return fmt.Errorf("remover diretório: %w", err)
			}
			if err := deps.Store.DeletePluginSource(name); err != nil {
				return err
			}
			if err := RunSync(deps, nil); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("plugin %q removido", name)))
			return nil
		},
	}
}

func confirmRemove(cmd *cobra.Command, name string) (bool, error) {
	fmt.Fprintf(cmd.ErrOrStderr(), "Tem certeza que deseja remover o plugin %q? (y/N): ", name)
	var answer string
	_, err := fmt.Fscanln(cmd.InOrStdin(), &answer)
	if err != nil && err.Error() != "unexpected newline" {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func newPluginsUpdateCmd(deps Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Atualiza um plugin ou todos (--all)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if all {
				sources, err := deps.Store.ListPluginSources()
				if err != nil {
					return err
				}
				for _, src := range sources {
					if src.GitURL == "" {
						continue
					}
					if err := updateOnePlugin(ctx, deps, src.InstallDir, cmd); err != nil {
						fmt.Fprintln(cmd.ErrOrStderr(), ui.RenderError(fmt.Sprintf("%s: %v", src.InstallDir, err)))
					}
				}
				return RunSync(deps, nil)
			}

			if len(args) == 0 {
				return fmt.Errorf("informe o nome do plugin ou use --all")
			}
			name := strings.TrimSpace(args[0])
			if err := updateOnePlugin(ctx, deps, name, cmd); err != nil {
				return err
			}
			return RunSync(deps, func(msg string) {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(msg))
			})
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Atualiza todos os plugins que tiverem nova versão")
	return cmd
}

func updateOnePlugin(ctx context.Context, deps Dependencies, name string, cmd *cobra.Command) error {
	src, err := deps.Store.GetPluginSource(name)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("plugin %q não encontrado no registry", name)
	}
	if src.GitURL == "" {
		return fmt.Errorf("plugin %q foi instalado manualmente (sem URL Git); não é possível atualizar", name)
	}

	dir := filepath.Join(deps.Runtime.PluginsDir, name)
	if !plugins.IsGitRepo(dir) {
		return fmt.Errorf("%s não é um repositório git", dir)
	}

	if src.RefType == "tag" {
		if err := plugins.FetchTags(ctx, dir); err != nil {
			return err
		}
		tags, err := plugins.ListLocalTags(dir)
		if err != nil {
			return err
		}
		var newerTag string
		for _, t := range tags {
			if _, isNewer := plugins.NewerTag(src.Ref, t); isNewer {
				newerTag = t
				break
			}
		}
		if newerTag == "" {
			if cmd != nil {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderInfo(fmt.Sprintf("%s: já está na versão mais recente (%s)", name, src.Ref)))
			}
			return nil
		}
		if err := plugins.CheckoutTag(ctx, dir, newerTag); err != nil {
			return err
		}
		version, _ := plugins.GetVersion(dir)
		src.Ref = newerTag
		src.Version = version
		if err := deps.Store.UpsertPluginSource(*src); err != nil {
			return err
		}
		if cmd != nil {
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("%s atualizado para %s", name, version)))
		}
		return nil
	}

	// branch
	if err := plugins.FetchAndPull(ctx, dir, src.Ref); err != nil {
		return err
	}
	version, err := plugins.GetVersion(dir)
	if err != nil {
		return err
	}
	src.Version = version
	if err := deps.Store.UpsertPluginSource(*src); err != nil {
		return err
	}
	if cmd != nil {
		fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("%s atualizado para %s", name, version)))
	}
	return nil
}
