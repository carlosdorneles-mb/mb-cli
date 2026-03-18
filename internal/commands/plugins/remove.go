package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/commands/self"
	"mb/internal/ui"
)

func newPluginsRemoveCmd(deps deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove um plugin instalado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			src, err := deps.Store.GetPluginSource(name)
			if err != nil {
				return err
			}
			if src == nil {
				return fmt.Errorf("plugin %q não encontrado", name)
			}

			confirmed, err := confirmRemove(cmd, name)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderInfo("remoção cancelada"))
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
			if err := self.RunSync(deps, nil, cmd.ErrOrStderr()); err != nil {
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
