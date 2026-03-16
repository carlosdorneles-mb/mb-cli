package plugincmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/commands/config"
	"mb/internal/commands/self"
	mbplugins "mb/internal/plugins"
	"mb/internal/ui"
)

func newPluginsUpdateCmd(deps config.Dependencies) *cobra.Command {
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
				return self.RunSync(deps, nil)
			}

			if len(args) == 0 {
				return fmt.Errorf("informe o nome do plugin ou use --all")
			}
			name := strings.TrimSpace(args[0])
			if err := updateOnePlugin(ctx, deps, name, cmd); err != nil {
				return err
			}
			return self.RunSync(deps, func(msg string) {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(msg))
			})
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Atualiza todos os plugins que tiverem nova versão")
	return cmd
}

func updateOnePlugin(ctx context.Context, deps config.Dependencies, name string, cmd *cobra.Command) error {
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
	if !mbplugins.IsGitRepo(dir) {
		return fmt.Errorf("%s não é um repositório git", dir)
	}

	if src.RefType == "tag" {
		if err := mbplugins.FetchTags(ctx, dir); err != nil {
			return err
		}
		tags, err := mbplugins.ListLocalTags(dir)
		if err != nil {
			return err
		}
		var newerTag string
		for _, t := range tags {
			if _, isNewer := mbplugins.NewerTag(src.Ref, t); isNewer {
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
		if err := mbplugins.CheckoutTag(ctx, dir, newerTag); err != nil {
			return err
		}
		version, _ := mbplugins.GetVersion(dir)
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

	if err := mbplugins.FetchAndPull(ctx, dir, src.Ref); err != nil {
		return err
	}
	version, err := mbplugins.GetVersion(dir)
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
