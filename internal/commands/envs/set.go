package envs

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/keyring"
	"mb/internal/system"
)

func newSetCmd(d deps.Dependencies) *cobra.Command {
	var setGroup string
	var secret bool
	cmd := &cobra.Command{
		Use:     "set <KEY> <VALUE>",
		Aliases: []string{"s"},
		Short:   "Define ou atualiza uma variável padrão ou pra um grupo específico",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			key, value := args[0], args[1]
			path, err := envTargetPath(d, setGroup)
			if err != nil {
				return err
			}
			group := envGroupForKeyring(setGroup)

			values, err := deps.LoadDefaultEnvValues(path)
			if err != nil {
				return err
			}

			if secret {
				if err := keyring.Set(group, key, value); err != nil {
					return err
				}
				if err := deps.AddSecretKey(path, key); err != nil {
					return err
				}
				delete(values, key)
				if err := deps.SaveDefaultEnvValues(path, values); err != nil {
					return err
				}
			} else {
				secretKeys, err := deps.LoadSecretKeys(path)
				if err != nil {
					return err
				}
				for _, sk := range secretKeys {
					if sk == key {
						_ = keyring.Delete(group, key)
						_ = deps.RemoveSecretKey(path, key)
						break
					}
				}
				values[key] = value
				if err := deps.SaveDefaultEnvValues(path, values); err != nil {
					return err
				}
			}

			if setGroup != "" {
				_ = log.Info(ctx, "variável %q salva no grupo %q", key, setGroup)
			} else {
				_ = log.Info(ctx, "variável %q salva no grupo padrão", key)
			}
			return nil
		},
	}
	cmd.Flags().
		StringVar(&setGroup, "group", "", "Grava a variável no grupo informado ao invés do grupo padrão")
	cmd.Flags().
		BoolVar(&secret, "secret", false, "Guarda o valor no keyring do sistema em vez do ficheiro env")
	cmd.GroupID = "commands"
	return cmd
}
