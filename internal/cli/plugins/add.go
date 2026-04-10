package plugins

import (
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	"mb/internal/usecase/addplugin"
)

type AddPluginService = addplugin.Service

func newPluginsAddCmd(svc *AddPluginService, d deps.Dependencies) *cobra.Command {
	var pkg string
	var tag string

	cmd := &cobra.Command{
		Use:     "add <git-url|path|.>",
		Aliases: []string{"install", "i", "a"},
		Short:   "Instala um plugin a partir de uma URL Git ou de um diretório local",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			return svc.Add(cmd.Context(), addplugin.Request{
				Source:  strings.TrimSpace(args[0]),
				Package: pkg,
				Tag:     tag,
			}, log)
		},
	}

	cmd.Flags().
		StringVar(&pkg, "package", "", "Identificador do pacote. Se não informado, usa o nome do repositório ou do diretório.")
	cmd.Flags().
		StringVar(&tag, "tag", "", "Instalar uma tag específica (apenas para plugin remoto).")
	return cmd
}
