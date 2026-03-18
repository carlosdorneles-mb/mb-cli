package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/commands/self"
	"mb/internal/system"
)

func newPluginsRemoveCmd(deps deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Aliases: []string{"rm", "r", "delete", "d", "del"},
		Short: "Remove um plugin instalado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(deps.Runtime.Quiet, deps.Runtime.Verbose, cmd.ErrOrStderr())
			name := strings.TrimSpace(args[0])
			src, err := deps.Store.GetPluginSource(name)
			if err != nil {
				return err
			}
			if src == nil {
				return fmt.Errorf("plugin %q não encontrado", name)
			}

			prompt := fmt.Sprintf("Remover o plugin %q?", name)
			confirmed, err := system.Confirm(ctx, prompt, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !confirmed {
				_ = log.Info(ctx, "remoção cancelada")
				return nil
			}

			if src.LocalPath == "" {
				destDir := filepath.Join(deps.Runtime.PluginsDir, name)
				if err := os.RemoveAll(destDir); err != nil {
					return fmt.Errorf("remover diretório: %w", err)
				}
			}
			if err := deps.Store.DeletePluginSource(name); err != nil {
				return err
			}
			if err := self.RunSync(ctx, deps, log, false); err != nil {
				return err
			}
			_ = log.Info(ctx, "plugin %q removido", name)
			return nil
		},
	}
}
