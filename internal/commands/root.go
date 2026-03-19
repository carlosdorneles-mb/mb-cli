package commands

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/browser"
	"mb/internal/commands/plugins"
	"mb/internal/commands/self"
	"mb/internal/config"
	"mb/internal/deps"
	"mb/internal/env"
	"mb/internal/plugincmd"
	"mb/internal/ui"
	"mb/internal/version"
)

type RootCommand = *cobra.Command

func docsURLForRuntime(d deps.Dependencies) string {
	u := strings.TrimSpace(d.AppConfig.DocsBaseURL)
	if u != "" {
		return u
	}
	return config.DefaultDocsURL
}

func NewRootCmd(d deps.Dependencies) RootCommand {
	var openDoc bool
	rootCmd := &cobra.Command{
		Use:   "mb",
		Short: "MB CLI - Uma CLI, infinitas possibilidades",
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
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if _, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues); err != nil {
				return err
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
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().
		BoolVarP(&d.Runtime.Verbose, "verbose", "v", false, "Ativa logs verbosos")
	rootCmd.PersistentFlags().
		BoolVarP(&d.Runtime.Quiet, "quiet", "q", false, "Não exibir nenhuma mensagem")
	rootCmd.PersistentFlags().
		StringVar(&d.Runtime.EnvFilePath, "env-file", "", "Caminho do arquivo .env")
	rootCmd.PersistentFlags().
		StringVar(&d.Runtime.EnvGroup, "env-group", "", "Carrega as váriaveis de um grupo específico")
	rootCmd.PersistentFlags().
		StringArrayVarP(&d.Runtime.InlineEnvValues, "env", "e", nil, "Define variável KEY=VALUE")
	rootCmd.Flags().BoolVar(&openDoc, "doc", false, "Abre a documentação no navegador")

	rootCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMMANDOS"})
	rootCmd.AddGroup(&cobra.Group{ID: "plugin_commands", Title: "COMMANDOS DE PLUGINS"})

	rootCmd.SetHelpCommandGroupID("commands")

	rootCmd.AddCommand(self.NewSelfCmd(d))
	rootCmd.AddCommand(plugins.NewPluginsCmd(d))
	plugincmd.Attach(rootCmd, d)

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
