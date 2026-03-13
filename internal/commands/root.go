package commands

import (
	"fmt"
	"os"

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

	rootCmd.PersistentFlags().BoolVar(&deps.Runtime.Verbose, "verbose", false, "enable verbose logs")
	rootCmd.PersistentFlags().BoolVar(&deps.Runtime.Quiet, "quiet", false, "disable banner and reduced output")
	rootCmd.PersistentFlags().StringVar(&deps.Runtime.EnvFilePath, "env-file", "", "path to .env file")
	rootCmd.PersistentFlags().StringArrayVar(&deps.Runtime.InlineEnvValues, "env", nil, "set env with KEY=VALUE")

	rootCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMMANDS"})
	rootCmd.AddGroup(&cobra.Group{ID: "plugin_commands", Title: "PLUGIN COMMANDS"})
	
	rootCmd.SetHelpCommandGroupID("commands")
	rootCmd.SetCompletionCommandGroupID("commands")

	rootCmd.AddCommand(NewSelfCmd(deps))
	AttachDynamicCommands(rootCmd, deps)

	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return rootCmd
}
