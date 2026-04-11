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
		Long: `Instala um plugin a partir de uma URL Git ou de um diretório local.

O MB detecta automaticamente se os plugins estão num subdiretório (default: src/).
Para usar outro subdiretório, defina MB_PLUGIN_SUBDIR. Para desativar, defina
MB_PLUGIN_SUBDIR= (vazio) e o scan será feito na raiz do pacote.

Quando --package não é informado, o nome do pacote é:
  - Remoto (Git): último segmento da URL (ex.: org/repo → repo)
  - Local: nome do diretório (ex.: /path/meu-plugin → meu-plugin)
Esse nome é usado em mb plugins remove <pacote> e mb plugins update <pacote>.`,
		Example: `# Instalar a partir do GitHub
  mb plugins add https://github.com/org/repo
  mb plugins add https://github.com/org/repo --tag v1.2.0 --package meu-plugin

  # Instalar a partir de um path local
  mb plugins add /caminho/para/meu-pacote
  mb plugins add . --package meu-plugin

  # Instalar com subdiretório personalizado (em vez de src/)
  MB_PLUGIN_SUBDIR=lib mb plugins add https://github.com/org/repo
  MB_PLUGIN_SUBDIR=plugins mb plugins add /caminho/para/meu-pacote`,
		Args: cobra.ExactArgs(1),
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
