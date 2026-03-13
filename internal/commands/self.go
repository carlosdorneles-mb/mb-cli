package commands

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"mb/internal/system"
	"mb/internal/ui"
)

func NewSelfCmd(deps Dependencies) *cobra.Command {
	selfCmd := &cobra.Command{
		Use:     "self",
		Short:   "Gerencia operações internas do MB CLI",
		GroupID: "commands",
	}

	selfCmd.AddCommand(newSelfSyncCmd(deps))
	selfCmd.AddCommand(newSelfListCmd(deps))
	selfCmd.AddCommand(newSelfEnvCmd(deps))
	return selfCmd
}

func newSelfSyncCmd(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Rescaneia plugins e reconstrói o cache SQLite",
		RunE: func(cmd *cobra.Command, _ []string) error {
			plugins, categories, err := deps.Scanner.Scan()
			if err != nil {
				return err
			}

			for _, plugin := range plugins {
				if err := deps.Store.UpsertPlugin(plugin); err != nil {
					return err
				}
			}

			if err := deps.Store.DeleteAllCategories(); err != nil {
				return err
			}
			for _, cat := range categories {
				if err := deps.Store.UpsertCategory(cat); err != nil {
					return err
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("synced %d plugin(s)", len(plugins))))
			return nil
		},
	}
}

func newSelfListCmd(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista os comandos disponíveis no cache",
		RunE: func(_ *cobra.Command, _ []string) error {
			plugins, err := deps.Store.ListPlugins()
			if err != nil {
				return err
			}

			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].CommandPath < plugins[j].CommandPath
			})

			rows := make([][]string, 0, len(plugins))
			for _, plugin := range plugins {
				kind := plugin.PluginType
				execOrFlags := plugin.ExecPath
				if plugin.FlagsJSON != "" {
					kind = "flags"
					execOrFlags = "(see --help)"
				}
				if execOrFlags == "" {
					execOrFlags = "-"
				}
				rows = append(rows, []string{
					plugin.CommandPath,
					plugin.CommandName,
					kind,
					execOrFlags,
				})
			}

			return system.Table(context.Background(), []string{"PATH", "COMMAND", "TYPE", "EXEC"}, rows)
		},
	}
}

func newSelfEnvCmd(deps Dependencies) *cobra.Command {
	selfEnvCmd := &cobra.Command{
		Use:   "env",
		Short: "Gerencia variáveis de ambiente padrão",
	}

	selfEnvCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Lista variáveis padrão",
		RunE: func(cmd *cobra.Command, _ []string) error {
			values, err := loadDefaultEnvValues(deps.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			keys := make([]string, 0, len(values))
			for key := range values {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", key, values[key])
			}
			return nil
		},
	})

	selfEnvCmd.AddCommand(&cobra.Command{
		Use:   "set KEY [VALUE]",
		Short: "Define uma variável padrão",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			var value string
			if len(args) == 2 {
				value = args[1]
			} else {
				input, err := system.Input(context.Background(), fmt.Sprintf("%s=", key))
				if err != nil {
					return err
				}
				value = input
			}

			values, err := loadDefaultEnvValues(deps.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			values[key] = value
			if err := saveDefaultEnvValues(deps.Runtime.DefaultEnvPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("saved %s", key)))
			return nil
		},
	})

	selfEnvCmd.AddCommand(&cobra.Command{
		Use:   "unset KEY",
		Short: "Remove uma variável padrão",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := loadDefaultEnvValues(deps.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			delete(values, args[0])
			if err := saveDefaultEnvValues(deps.Runtime.DefaultEnvPath, values); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("removed %s", args[0])))
			return nil
		},
	})

	return selfEnvCmd
}
