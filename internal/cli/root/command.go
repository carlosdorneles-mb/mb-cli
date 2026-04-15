package root

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cli/alias"
	"mb/internal/cli/completion"
	"mb/internal/cli/envs"
	"mb/internal/cli/plugincmd"
	"mb/internal/cli/plugins"
	"mb/internal/cli/run"
	"mb/internal/cli/runtimeflags"
	"mb/internal/cli/update"
	"mb/internal/deps"
	"mb/internal/infra/browser"
	"mb/internal/infra/updatecheck"
	"mb/internal/ports"
	"mb/internal/shared/config"
	"mb/internal/shared/env"
	"mb/internal/shared/ui"
	"mb/internal/shared/version"
	"mb/internal/usecase/addplugin"
	usecaseenvs "mb/internal/usecase/envs"
	usecaseplugins "mb/internal/usecase/plugins"
)

type RootCommand = *cobra.Command

func docsURLForRuntime(d deps.Dependencies) string {
	u := strings.TrimSpace(d.AppConfig.DocsBaseURL)
	if u != "" {
		return u
	}
	return config.DefaultDocsURL
}

// shouldSkipUpdateCheck returns true when the command doesn't warrant an update check.
func shouldSkipUpdateCheck(cmd *cobra.Command) bool {
	// Skip if disabled via env var
	if updatecheck.IsDisabled() {
		return true
	}

	// Skip for non-release builds
	if !updatecheck.IsReleaseBuild() {
		return true
	}

	// Get the command path (e.g., "update", "completion install")
	cmdPath := cmd.CommandPath()

	// Skip for update command itself
	if strings.HasPrefix(cmdPath, "mb update") {
		return true
	}

	// Skip for help and version
	if cmd.Name() == "help" {
		return true
	}
	if cmd.Flags().Lookup("help") != nil && cmd.Flags().Lookup("help").Changed {
		return true
	}
	if cmd.Flags().Lookup("version") != nil && cmd.Flags().Lookup("version").Changed {
		return true
	}

	// Skip for completion subcommands
	if strings.HasPrefix(cmdPath, "mb completion") {
		return true
	}

	return false
}

func NewRootCmd(
	d deps.Dependencies,
	fsys ports.Filesystem,
	git ports.GitOperations,
	shell ports.ShellHelperInstaller,
	layout ports.PluginLayoutValidator,
) RootCommand {
	addPluginSvc, syncSvc, rmSvc, upSvc, listSvc := buildPluginServices(d, fsys, git, shell, layout)
	var openDoc bool
	rootCmd := &cobra.Command{
		Use:   "mb",
		Short: "MB CLI é uma ferramenta CLI que transforma plugins em comandos dinâmicos, com cache, injeção segura de variáveis de ambiente e helpers de shell poderosos.",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if !openDoc {
				return nil
			}
			docURL := docsURLForRuntime(d)
			if err := browser.OpenURL(docURL); err != nil {
				return err
			}
			if !d.Runtime.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Documentação: %s\n", docURL)
			}
			os.Exit(0)
			return nil
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues); err != nil {
				return err
			}

			// Version check: skip for update, help, version commands
			if !shouldSkipUpdateCheck(cmd) {
				checker := updatecheck.NewChecker(
					d.Runtime.ConfigDir,
					version.Version,
				)
				// Non-blocking: errors are silently ignored
				_ = checker.Run(cmd.Context())

				// Show warning if update available
				if tag, ok := checker.IsUpdateAvailable(); ok {
					fmt.Fprintf(cmd.ErrOrStderr(),
						"\n⚠️  Nova versão disponível: %s (atual: %s)\n"+
							"Execute: mb update --only-cli\n\n", tag, version.Version)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !d.Runtime.Quiet {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderBanner(ui.Banner))
			}
			cmd.Help()
			return nil
		},
		SilenceUsage: true,
	}
	runtimeflags.RegisterRuntimePersistentFlags(rootCmd.PersistentFlags(), d.Runtime)
	rootCmd.Flags().BoolVar(&openDoc, "doc", false, "Abre a documentação no navegador")

	rootCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})
	rootCmd.AddGroup(&cobra.Group{ID: "plugin_commands", Title: "PLUGINS"})

	rootCmd.SetHelpCommandGroupID("commands")

	envsCmd := envs.NewCmd(listSvc, d)
	envsCmd.GroupID = "commands"
	rootCmd.AddCommand(envsCmd)

	runCmd := run.NewRunCmd(d)
	runCmd.GroupID = "commands"
	rootCmd.AddCommand(runCmd)

	aliasCmd := alias.NewCmd(d)
	aliasCmd.GroupID = "commands"
	rootCmd.AddCommand(aliasCmd)

	pluginsCmd := plugins.NewPluginsCmd(addPluginSvc, syncSvc, rmSvc, upSvc, d)
	pluginsCmd.GroupID = "commands"
	rootCmd.AddCommand(pluginsCmd)

	updateCmd := update.NewUpdateCmd(upSvc, d)
	updateCmd.GroupID = "commands"
	rootCmd.AddCommand(updateCmd)

	plugincmd.Attach(rootCmd, d)

	rootCmd.InitDefaultCompletionCmd()
	customizeCompletionPT(rootCmd)
	customizeAliasPT(rootCmd)

	rootCmd.InitDefaultHelpCmd()
	for _, c := range rootCmd.Commands() {
		if c.Name() == "help" {
			c.Short = "Ajuda sobre qualquer commando"
			break
		}
	}

	if rootCmd.Version == "" {
		rootCmd.Version = "dev"
		if version.Version != "" {
			rootCmd.Version = version.Version
		} else if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			rootCmd.Version = info.Main.Version
		}
	}
	rootCmd.InitDefaultHelpFlag()
	if rootCmd.Version != "" {
		rootCmd.Flags().BoolP("version", "V", false, "Versão do MB CLI")
		_ = rootCmd.Flags().
			SetAnnotation("version", cobra.FlagSetByCobraAnnotation, []string{"true"})
	}
	rootCmd.InitDefaultVersionFlag()
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	initDefaultHelpFlagRecursive(rootCmd)
	setHelpFlagUsagePT(rootCmd)
	if f := rootCmd.Flags().Lookup("help"); f != nil {
		f.Usage = "Ajuda para MB CLI"
	}

	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return rootCmd
}

func initDefaultHelpFlagRecursive(cmd *cobra.Command) {
	cmd.InitDefaultHelpFlag()
	for _, sub := range cmd.Commands() {
		initDefaultHelpFlagRecursive(sub)
	}
}

func setHelpFlagUsagePT(cmd *cobra.Command) {
	if f := cmd.Flags().Lookup("help"); f != nil {
		displayName := cmd.DisplayName()
		if displayName == "" {
			displayName = "este commando"
		}
		f.Usage = "Ajuda para " + displayName
	}
	for _, sub := range cmd.Commands() {
		setHelpFlagUsagePT(sub)
	}
}

func findCommand(cmds []*cobra.Command, name string) *cobra.Command {
	for _, c := range cmds {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func customizeAliasPT(rootCmd *cobra.Command) {
	aliasCmd := findCommand(rootCmd.Commands(), "alias")
	if aliasCmd == nil {
		return
	}
	shortPT := map[string]string{
		"set":   "Define ou atualiza um alias",
		"list":  "Lista aliases (fzf e preview no TTY; tabela em pipe; --json para jq)",
		"unset": "Remove um ou mais aliases registrados",
	}
	for _, sub := range aliasCmd.Commands() {
		if short, ok := shortPT[sub.Name()]; ok {
			sub.Short = short
		}
	}
	if f := aliasCmd.PersistentFlags().Lookup("shell"); f != nil {
		f.Usage = "Shell alvo (bash, zsh, fish, powershell); por padrão detecta via SHELL"
	}
	if setCmd := findCommand(aliasCmd.Commands(), "set"); setCmd != nil {
		if f := setCmd.Flags().Lookup("yes"); f != nil {
			f.Usage = "Confirma alterações a aliases existentes sem prompt (CI / não interativo)"
		}
	}
	if unsetCmd := findCommand(aliasCmd.Commands(), "unset"); unsetCmd != nil {
		if f := unsetCmd.Flags().Lookup("yes"); f != nil {
			f.Usage = "Confirma a remoção sem prompt (CI / não interativo)"
		}
	}
}

func customizeCompletionPT(rootCmd *cobra.Command) {
	completionCmd := findCommand(rootCmd.Commands(), "completion")
	if completionCmd == nil {
		return
	}
	completionCmd.Short = "Gera o script de autocompletar do shell"
	completionCmd.GroupID = "commands"
	completionCmd.Long = "Gera o script de autocompletar para o MB CLI para o shell especificado.\nConsulte a ajuda de cada subcomando para detalhes de como usar o script gerado."
	const completionGroupID = "completion_shells"
	completionCmd.AddGroup(&cobra.Group{ID: completionGroupID, Title: "COMANDOS"})
	completionCmd.AddCommand(completion.NewInstallCmd(rootCmd))
	completionCmd.AddCommand(completion.NewUninstallCmd())
	shortPT := map[string]string{
		"bash":       "Gera o script de autocompletar para bash",
		"zsh":        "Gera o script de autocompletar para zsh",
		"fish":       "Gera o script de autocompletar para fish",
		"powershell": "Gera o script de autocompletar para powershell",
		"install":    "Instala ou atualiza o autocompletar no ficheiro de perfil do shell",
		"uninstall":  "Remove o bloco de autocompletar mb-cli do ficheiro de perfil",
	}
	for _, sub := range completionCmd.Commands() {
		sub.GroupID = completionGroupID
		if short, ok := shortPT[sub.Name()]; ok {
			sub.Short = short
		}
		if sub.Name() == "install" {
			sub.Long = `Detecta o shell (variável SHELL) ou usa --shell, gera o mesmo script que
«mb completion <shell>» e grava um bloco idempotente no ficheiro de perfil
(.bashrc, .zshrc, fish/config.fish ou perfil PowerShell) ou em --rc-file.

Em ambientes não interativos é obrigatório --yes (ou use --dry-run para pré-visualizar).`
		}
		if sub.Name() == "uninstall" {
			sub.Long = `Remove o bloco de autocompletar delimitado pelos marcadores mb-cli do ficheiro
de perfil por omissão do shell (ou de --rc-file). Não altera nada se o ficheiro
ou o bloco não existire.

Em ambientes não interativos é obrigatório --yes (ou use --dry-run para pré-visualizar).`
		}
		if f := sub.Flags().Lookup("no-descriptions"); f != nil {
			f.Usage = "Desativa as descrições no autocompletar"
		}
	}
}

func buildPluginServices(
	d deps.Dependencies,
	fsys ports.Filesystem,
	git ports.GitOperations,
	shell ports.ShellHelperInstaller,
	layout ports.PluginLayoutValidator,
) (*addplugin.Service, *usecaseplugins.SyncService, *usecaseplugins.RemoveService, *usecaseplugins.UpdateService, *usecaseenvs.ListService) {
	rt := usecaseplugins.PluginRuntime{
		ConfigDir:  d.Runtime.ConfigDir,
		PluginsDir: d.Runtime.PluginsDir,
		Quiet:      d.Runtime.Quiet,
		Verbose:    d.Runtime.Verbose,
	}

	syncer := addplugin.NewSyncer()
	syncSvc := usecaseplugins.NewSyncService(rt, d.Store, d.Scanner, shell)

	addPluginSvc := addplugin.New(
		addplugin.Runtime{ConfigDir: d.Runtime.ConfigDir, PluginsDir: d.Runtime.PluginsDir},
		d.Store,
		d.Scanner,
		fsys,
		git,
		shell,
		layout,
		syncer,
	)

	rmSvc := usecaseplugins.NewRemoveService(rt, d.Store, d.Scanner, shell, fsys, syncSvc)
	upSvc := usecaseplugins.NewUpdateService(rt, d.Store, d.Scanner, shell, git, fsys, syncSvc)

	paths := usecaseenvs.Paths{
		DefaultEnvPath: d.Runtime.DefaultEnvPath,
		ConfigDir:      d.Runtime.ConfigDir,
	}
	listSvc := usecaseenvs.NewListService(d.SecretStore, d.OnePassword, paths)

	return addPluginSvc, syncSvc, rmSvc, upSvc, listSvc
}
