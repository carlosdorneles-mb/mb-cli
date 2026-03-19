package plugins

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	mbplugins "mb/internal/infra/plugins"
	"mb/internal/shared/system"
)

// RunUpdateAll updates all plugins that have a GitURL and no LocalPath, then runs sync.
func RunUpdateAll(ctx context.Context, deps deps.Dependencies, log *system.Logger) error {
	sources, err := deps.Store.ListPluginSources()
	if err != nil {
		return err
	}
	for _, src := range sources {
		if src.GitURL == "" || src.LocalPath != "" {
			continue
		}
		if err := updateOnePlugin(ctx, deps, log, src.InstallDir, nil); err != nil {
			_ = log.Error(ctx, "%s: %v", src.InstallDir, err)
		}
	}
	return RunSync(ctx, deps, log, false)
}

func newPluginsUpdateCmd(deps deps.Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:     "update <name>",
		Aliases: []string{"up", "u"},
		Short:   "Atualiza um plugin ou todos (--all)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())

			if all {
				return RunUpdateAll(ctx, deps, log)
			}

			if len(args) == 0 {
				return fmt.Errorf("informe o nome do plugin ou use --all")
			}
			name := strings.TrimSpace(args[0])
			if err := updateOnePlugin(ctx, deps, log, name, cmd); err != nil {
				return err
			}
			return RunSync(ctx, deps, log, true)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Atualiza todos os plugins que tiverem nova versão")
	return cmd
}

func updateOnePlugin(
	ctx context.Context,
	deps deps.Dependencies,
	log *system.Logger,
	name string,
	cmd *cobra.Command,
) error {
	src, err := deps.Store.GetPluginSource(name)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("plugin %q não encontrado no registry", name)
	}
	if src.LocalPath != "" {
		return fmt.Errorf("plugin %q é local; não é possível atualizar", name)
	}
	if src.GitURL == "" {
		return fmt.Errorf(
			"plugin %q foi instalado manualmente (sem URL Git); não é possível atualizar",
			name,
		)
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
			if cmd != nil && log != nil {
				_ = log.Info(ctx, "%s: já está na versão mais recente (%s)", name, src.Ref)
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
		if cmd != nil && log != nil {
			_ = log.Info(ctx, "%s atualizado para %s", name, version)
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
	if cmd != nil && log != nil {
		_ = log.Info(ctx, "%s atualizado para %s", name, version)
	}
	return nil
}
