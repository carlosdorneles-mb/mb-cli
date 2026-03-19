package envs

import (
	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/keyring"
	"mb/internal/system"
)

func newUnsetCmd(d deps.Dependencies) *cobra.Command {
	var unsetGroup string
	cmd := &cobra.Command{
		Use:     "unset <KEY>",
		Aliases: []string{"u"},
		Short:   "Remove uma variável padrão ou de um grupo específico",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			path, err := envTargetPath(d, unsetGroup)
			if err != nil {
				return err
			}
			group := envGroupForKeyring(unsetGroup)

			values, err := deps.LoadDefaultEnvValues(path)
			if err != nil {
				return err
			}

			key := args[0]
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
			delete(values, key)
			if err := deps.SaveDefaultEnvValues(path, values); err != nil {
				return err
			}
			if unsetGroup != "" {
				_ = log.Info(ctx, "variável %q removida do grupo %q", key, unsetGroup)
			} else {
				_ = log.Info(ctx, "variável %q removida do grupo padrão", key)
			}
			return nil
		},
	}
	cmd.Flags().
		StringVar(&unsetGroup, "group", "", "Remove do arquivo referente ao grupo informado")
	cmd.GroupID = "commands"
	return cmd
}
