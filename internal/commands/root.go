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
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().BoolVarP(&deps.Runtime.Verbose, "verbose", "v", false, "Ativa logs verbosos")
	rootCmd.PersistentFlags().BoolVarP(&deps.Runtime.Quiet, "quiet", "q", false, "Não exibir nenhuma mensagem")
	rootCmd.PersistentFlags().StringVar(&deps.Runtime.EnvFilePath, "env-file", "", "Caminho do arquivo .env")
	rootCmd.PersistentFlags().StringArrayVarP(&deps.Runtime.InlineEnvValues, "env", "e", nil, "Define variável KEY=VALUE")

	rootCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})
	rootCmd.AddGroup(&cobra.Group{ID: "plugin_commands", Title: "COMANDOS DE PLUGINS"})
	
	rootCmd.SetHelpCommandGroupID("commands")

	rootCmd.AddCommand(self.NewSelfCmd(deps))
	rootCmd.AddCommand(plugincmd.NewPluginsCmd(deps))
	AttachDynamicCommands(rootCmd, deps)

	rootCmd.InitDefaultHelpCmd()
	for _, c := range rootCmd.Commands() {
		if c.Name() == "help" {
			c.Short = "Ajuda sobre qualquer comando"
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
		_ = rootCmd.Flags().SetAnnotation("version", cobra.FlagSetByCobraAnnotation, []string{"true"})
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
			displayName = "este comando"
		}
		f.Usage = "Ajuda para " + displayName
	}
	for _, sub := range cmd.Commands() {
		setHelpFlagUsagePT(sub)
	}
}
