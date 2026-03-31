package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	appplugins "mb/internal/app/plugins"
	"mb/internal/deps"
	mbfs "mb/internal/infra/fs"
	mbplugins "mb/internal/infra/plugins"
	"mb/internal/infra/shellhelpers"
	"mb/internal/shared/system"
)

func newPluginsAddCmd(d deps.Dependencies) *cobra.Command {
	var pkg string
	var tag string
	var noRemove bool

	cmd := &cobra.Command{
		Use:     "add <git-url|path|.>",
		Aliases: []string{"install", "i", "a"},
		Short:   "Instala um plugin a partir de uma URL Git ou de um diretório local",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := strings.TrimSpace(args[0])
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			syncOpts := appplugins.SyncOptions{EmitSuccess: false, NoRemove: noRemove}
			syncOpts = withCompletionPostSync(cmd, d, log, syncOpts)
			_, _, err := mbplugins.ParseGitURL(arg)
			if err == nil {
				return runAddRemoteCLI(cmd, d, arg, pkg, tag, syncOpts)
			}
			return runAddLocalCLI(cmd, d, log, arg, pkg, syncOpts)
		},
	}

	cmd.Flags().
		StringVar(&pkg, "package", "", "Identificador do pacote. Se não informado, usa o nome do repositório ou do diretório.")
	cmd.Flags().
		StringVar(&tag, "tag", "", "Instalar uma tag específica (apenas para plugin remoto).")
	cmd.Flags().BoolVar(&noRemove, "no-remove", false,
		"Mantém no cache comandos removidos do pacote",
	)
	return cmd
}

func pluginRuntimeFromDeps(d deps.Dependencies) appplugins.PluginRuntime {
	return appplugins.PluginRuntime{
		ConfigDir:  d.Runtime.ConfigDir,
		PluginsDir: d.Runtime.PluginsDir,
		Quiet:      d.Runtime.Quiet,
		Verbose:    d.Runtime.Verbose,
	}
}

func runAddRemoteCLI(
	cmd *cobra.Command,
	d deps.Dependencies,
	gitURL string,
	pkg string,
	tag string,
	syncOpts appplugins.SyncOptions,
) error {
	ctx := cmd.Context()
	log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
	return appplugins.RunAddRemote(
		ctx,
		pluginRuntimeFromDeps(d),
		d.Store,
		d.Scanner,
		shellhelpers.Installer{},
		mbplugins.GitService{},
		mbfs.OS{},
		log,
		gitURL,
		pkg,
		tag,
		syncOpts,
	)
}

func runAddLocalCLI(
	cmd *cobra.Command,
	d deps.Dependencies,
	log *system.Logger,
	pathArg string,
	pkg string,
	syncOpts appplugins.SyncOptions,
) error {
	if pathArg == "" {
		return fmt.Errorf("informe a URL do repositório, um path ou . para o diretório atual")
	}
	var absPath string
	var err error
	if pathArg == "." {
		absPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("obter diretório atual: %w", err)
		}
	} else {
		absPath, err = filepath.Abs(pathArg)
		if err != nil {
			return fmt.Errorf("caminho inválido: %w", err)
		}
	}
	ctx := cmd.Context()
	return appplugins.RunAddLocalPath(
		ctx,
		pluginRuntimeFromDeps(d),
		d.Store,
		d.Scanner,
		shellhelpers.Installer{},
		mbfs.OS{},
		mbplugins.LayoutValidator{},
		log,
		absPath,
		pkg,
		syncOpts,
	)
}
