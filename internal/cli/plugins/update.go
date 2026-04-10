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
