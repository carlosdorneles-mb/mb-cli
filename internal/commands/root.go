package commands

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/env"
	"mb/internal/executor"
	"mb/internal/plugins"
	"mb/internal/ui"
)

type RootCommand = *cobra.Command

type RuntimeConfig struct {
	ConfigDir       string
	PluginsDir      string
	CacheDBPath     string
	DefaultEnvPath  string
	Verbose         bool
	Quiet           bool
	EnvFilePath     string
	InlineEnvValues []string
}

type Dependencies struct {
	Runtime  *RuntimeConfig
	Store    *cache.Store
	Scanner  *plugins.Scanner
	Executor *executor.Executor
}

func NewDependencies(
	runtime *RuntimeConfig,
	store *cache.Store,
	scanner *plugins.Scanner,
	exec *executor.Executor,
) Dependencies {
	return Dependencies{
		Runtime:  runtime,
		Store:    store,
		Scanner:  scanner,
		Executor: exec,
	}
}

func NewRootCmd(deps Dependencies) RootCommand {
	rootCmd := &cobra.Command{
		Use:   "mb",
		Short: "MB CLI - Orquestrador Moderno em Go",
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

	rootCmd.PersistentFlags().BoolVar(&deps.Runtime.Verbose, "verbose", false, "Ativa logs verbosos")
	rootCmd.PersistentFlags().BoolVar(&deps.Runtime.Quiet, "quiet", false, "Desativa banner e reduz saída")
	rootCmd.PersistentFlags().StringVar(&deps.Runtime.EnvFilePath, "env-file", "", "Caminho do arquivo .env")
	rootCmd.PersistentFlags().StringArrayVar(&deps.Runtime.InlineEnvValues, "env", nil, "Define variável KEY=VALUE")

	rootCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})
	rootCmd.AddGroup(&cobra.Group{ID: "plugin_commands", Title: "COMANDOS DE PLUGINS"})
	
	rootCmd.SetHelpCommandGroupID("commands")
	rootCmd.SetCompletionCommandGroupID("commands")

	rootCmd.AddCommand(NewSelfCmd(deps))
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
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			rootCmd.Version = info.Main.Version
		}
	}
	rootCmd.InitDefaultHelpFlag()
	rootCmd.InitDefaultVersionFlag()
	if f := rootCmd.Flags().Lookup("help"); f != nil {
		f.Usage = "Ajuda para MB CLI"
	}
	if f := rootCmd.Flags().Lookup("version"); f != nil {
		f.Usage = "Versão do MB CLI"
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
