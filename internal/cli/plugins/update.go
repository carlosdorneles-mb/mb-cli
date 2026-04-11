package plugins

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	"mb/internal/usecase/plugins"
)

func newPluginsUpdateCmd(svc *plugins.UpdateService, d deps.Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:     "update <package>",
		Aliases: []string{"up", "u"},
		Short:   "Atualiza um plugin ou todos (--all)",
		Long: `Atualiza um plugin instalado ou todos com --all.

O nome do pacote é o valor da coluna PACOTE em mb plugins list.
Ao instalar sem --package, usa-se o nome do repositório (Git) ou do diretório (local).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

			if all {
				return svc.Update(ctx, plugins.UpdateRequest{}, log)
			}

			if len(args) == 0 {
				return fmt.Errorf("informe o pacote ou use --all")
			}
			pkg := strings.TrimSpace(args[0])
			return svc.Update(ctx, plugins.UpdateRequest{Package: pkg}, log)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Atualiza todos os plugins que tiverem nova versão")
	return cmd
}
