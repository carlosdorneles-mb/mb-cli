package self

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/system"
	"mb/internal/ui"
)

func newSelfEnvCmd(d deps.Dependencies) *cobra.Command {
	selfEnvCmd := &cobra.Command{
		Use:   "env",
		Short: "Gerencia variáveis de ambiente padrão",
	}
	selfEnvCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista variáveis padrão",
		RunE: func(cmd *cobra.Command, _ []string) error {
			values, err := deps.LoadDefaultEnvValues(d.Runtime.DefaultEnvPath)
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
	}
	listCmd.GroupID = "commands"
	selfEnvCmd.AddCommand(listCmd)

	setCmd := &cobra.Command{
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

			values, err := deps.LoadDefaultEnvValues(d.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			values[key] = value
			if err := deps.SaveDefaultEnvValues(d.Runtime.DefaultEnvPath, values); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("salvo %s", key)))
			return nil
		},
	}
	setCmd.GroupID = "commands"
	selfEnvCmd.AddCommand(setCmd)

	unsetCmd := &cobra.Command{
		Use:   "unset KEY",
		Short: "Remove uma variável padrão",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := deps.LoadDefaultEnvValues(d.Runtime.DefaultEnvPath)
			if err != nil {
				return err
			}

			delete(values, args[0])
			if err := deps.SaveDefaultEnvValues(d.Runtime.DefaultEnvPath, values); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(fmt.Sprintf("removido %s", args[0])))
			return nil
		},
	}
	unsetCmd.GroupID = "commands"
	selfEnvCmd.AddCommand(unsetCmd)

	return selfEnvCmd
}
