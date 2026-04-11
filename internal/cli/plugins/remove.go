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
	var all bool

	cmd := &cobra.Command{
		Use:     "remove [<package>...]",
		Aliases: []string{"rm", "r", "delete", "d", "del"},
		Short:   "Remove um ou mais plugins instalados, ou todos com --all",
		Long: `Remove um ou mais plugins instalados pelo nome do pacote (coluna PACOTE em mb plugins list).

O nome do pacote é:
  - Remoto (Git): último segmento da URL ao adicionar (org/repo → repo)
  - Local: nome do diretório ao adicionar (/path/meu-plugin → meu-plugin)
  - Ou o valor informado com --package ao instalar`,
		Example: `# Remove apenas um pacote
  mb plugins remove infra-tools

  # Remove mais de um pacote
  mb plugins remove infra-tools deploy-scripts

  # Remove todos os plugins
  mb plugins remove --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

			var pkgs []string

			if all {
				sources, err := d.Store.ListPluginSources()
				if err != nil {
					return err
				}
				if len(sources) == 0 {
					_ = log.Info(ctx, "Não há plugins instalados")
					return nil
				}
				for _, s := range sources {
					pkgs = append(pkgs, s.InstallDir)
				}
			} else {
				if len(args) == 0 {
					return fmt.Errorf("informe o pacote ou use --all")
				}
				for _, a := range args {
					pkg := strings.TrimSpace(a)
					if pkg == "" {
						continue
					}
					src, err := d.Store.GetPluginSource(pkg)
					if err != nil {
						return err
					}
					if src == nil {
						return fmt.Errorf("pacote %q não encontrado (use mb plugins list)", pkg)
					}
					pkgs = append(pkgs, pkg)
				}
				if len(pkgs) == 0 {
					return fmt.Errorf("nenhum pacote válido informado")
				}
			}

			prompt := fmt.Sprintf(
				"Remover os %d pacote(s) %s?",
				len(pkgs),
				strings.Join(pkgs, ", "),
			)
			confirmed, err := system.Confirm(ctx, prompt, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !confirmed {
				_ = log.Info(ctx, "Remoção cancelada")
				return nil
			}

			for _, pkg := range pkgs {
				if err := svc.Remove(ctx, plugins.RemoveRequest{Package: pkg}, log); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Remove todos os plugins instalados")
	return cmd
}
