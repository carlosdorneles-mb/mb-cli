package env

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/gumlog"
)

func newSetCmd(d deps.Dependencies) *cobra.Command {
	var setGroup string
	cmd := &cobra.Command{
		Use:     "set <KEY> <VALUE>",
		Aliases: []string{"s"},
		Short:   "Define uma variável padrão ou pra um grupo específico",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := gumlog.New(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			key, value := args[0], args[1]
			path, err := envTargetPath(d, setGroup)
			if err != nil {
				return err
			}

			values, err := deps.LoadDefaultEnvValues(path)
			if err != nil {
				return err
			}

			values[key] = value
			if err := deps.SaveDefaultEnvValues(path, values); err != nil {
				return err
			}

			if setGroup != "" {
				_ = log.Info(ctx, "variável %q salva no grupo %q", key, setGroup)
			} else {
				_ = log.Info(ctx, "variável %q salva no grupo padrão", key)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&setGroup, "group", "", "Grava a variável no grupo informado ao invés do grupo padrão")
	cmd.GroupID = "commands"
	return cmd
}
