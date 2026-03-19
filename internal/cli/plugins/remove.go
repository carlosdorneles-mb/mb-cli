package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
)

func newPluginsRemoveCmd(deps deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <package>",
		Aliases: []string{"rm", "r", "delete", "d", "del"},
		Short:   "Remove um plugin instalado",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())
			pkg := strings.TrimSpace(args[0])
			src, err := deps.Store.GetPluginSource(pkg)
			if err != nil {
				return err
			}
			if src == nil {
				return fmt.Errorf("pacote %q não encontrado", pkg)
			}

			prompt := fmt.Sprintf("Remover o pacote %q?", pkg)
			confirmed, err := system.Confirm(ctx, prompt, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !confirmed {
				_ = log.Info(ctx, "remoção cancelada")
				return nil
			}

			if src.LocalPath == "" {
				destDir := filepath.Join(deps.Runtime.PluginsDir, pkg)
				if err := os.RemoveAll(destDir); err != nil {
					return fmt.Errorf("remover diretório: %w", err)
				}
			}
			if err := deps.Store.DeletePluginSource(pkg); err != nil {
				return err
			}
			if err := RunSync(ctx, deps, log, false); err != nil {
				return err
			}
			_ = log.Info(ctx, "pacote %q removido", pkg)
			return nil
		},
	}
}
