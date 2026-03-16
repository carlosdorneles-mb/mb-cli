package commands

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"mb/internal/commands/config"
	plugincmd "mb/internal/commands/plugins"
	"mb/internal/commands/self"
	"mb/internal/env"
	"mb/internal/ui"
	"mb/internal/version"
)

type RootCommand = *cobra.Command

func NewRootCmd(deps config.Dependencies) RootCommand {
	rootCmd := &cobra.Command{
		Use:   "mb",
		Short: "MB CLI - Uma CLI, infinitas possibilidades",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if _, err := env.ParseInlinePairs(deps.Runtime.InlineEnvValues); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !deps.Runtime.Quiet {
				fmt.Fprintln(cmd.OutOrStdout(), ui.RenderBanner(ui.Banner))
			}
			cmd.Help()
			return nil
		},
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&deps.Runtime.Verbose, "verbose", "v", false, "Ativa logs verbosos")
	rootCmd.PersistentFlags().BoolVarP(&deps.Runtime.Quiet, "quiet", "q", false, "Não exibir nenhuma mensagem")
	rootCmd.PersistentFlags().StringVar(&deps.Runtime.EnvFilePath, "env-file", "", "Caminho do arquivo .env")
	rootCmd.PersistentFlags().StringArrayVarP(&deps.Runtime.InlineEnvValues, "env", "e", nil, "Define variável KEY=VALUE")

	rootCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})
	rootCmd.AddGroup(&cobra.Group{ID: "plugin_commands", Title: "COMANDOS DE PLUGINS"})
	
	rootCmd.SetHelpCommandGroupID("commands")
	rootCmd.SetCompletionCommandGroupID("commands")

	rootCmd.AddCommand(self.NewSelfCmd(deps))
	rootCmd.AddCommand(plugincmd.NewPluginsCmd(deps))
	AttachDynamicCommands(rootCmd, deps)

	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	for _, c := range rootCmd.Commands() {
		switch c.Name() {
		case "help":
			c.Short = "Ajuda sobre qualquer comando"
		case "completion":
			c.Short = "Gera o script de autocompletar do shell"
		}
	}

	// Completion subcommands: group, Long, and Short in Portuguese
	if completionCmd := findCommand(rootCmd.Commands(), "completion"); completionCmd != nil {
		completionCmd.Long = "Gera o script de autocompletar para o MB CLI para o shell especificado.\nConsulte a ajuda de cada subcomando para detalhes de como usar o script gerado."
		const completionGroupID = "completion_shells"
		completionCmd.AddGroup(&cobra.Group{ID: completionGroupID, Title: "COMANDOS"})
		shortPT := map[string]string{
			"bash":       "Gera o script de autocompletar para bash",
			"zsh":        "Gera o script de autocompletar para zsh",
			"fish":       "Gera o script de autocompletar para fish",
			"powershell": "Gera o script de autocompletar para powershell",
		}
		for _, sub := range completionCmd.Commands() {
			sub.GroupID = completionGroupID
			if short, ok := shortPT[sub.Name()]; ok {
				sub.Short = short
			}
			if f := sub.Flags().Lookup("no-descriptions"); f != nil {
				f.Usage = "Desativa as descrições no autocompletar"
			}
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
		_ = rootCmd.Flags().SetAnnotation("version", cobra.FlagSetByCobraAnnotation, []string{"true"})
	}
	rootCmd.InitDefaultVersionFlag()
	initDefaultHelpFlagRecursive(rootCmd)
	setHelpFlagUsagePT(rootCmd)
	if f := rootCmd.Flags().Lookup("help"); f != nil {
		f.Usage = "Ajuda para MB CLI"
	}

	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return rootCmd
}

func findCommand(cmds []*cobra.Command, name string) *cobra.Command {
	for _, c := range cmds {
		if c.Name() == name {
			return c
		}
	}
	return nil
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
			displayName = "este comando"
		}
		f.Usage = "Ajuda para " + displayName
	}
	for _, sub := range cmd.Commands() {
		setHelpFlagUsagePT(sub)
	}
}
