package plugins

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	"mb/internal/usecase/plugins"
)

func newPluginsRemoveCmd(svc *plugins.RemoveService, d deps.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <package>",
		Aliases: []string{"rm", "r", "delete", "d", "del"},
		Short:   "Remove um plugin instalado",
		Long: `Remove um plugin instalado pelo nome do pacote (coluna PACOTE em mb plugins list).

O nome do pacote é:
  - Remoto (Git): último segmento da URL ao adicionar (org/repo → repo)
  - Local: nome do diretório ao adicionar (/path/meu-plugin → meu-plugin)
  - Ou o valor informado com --package ao instalar`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			pkg := strings.TrimSpace(args[0])

			src, err := d.Store.GetPluginSource(pkg)
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
				_ = log.Info(ctx, "Remoção cancelada")
				return nil
			}

			return svc.Remove(ctx, plugins.RemoveRequest{Package: pkg}, log)
		},
	}
}
