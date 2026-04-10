package envs

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/deps"
	"mb/internal/shared/system"
	appenvs "mb/internal/usecase/envs"
)

func newSetCmd(d deps.Dependencies) *cobra.Command {
	var setVault string
	var flagSecret, flagSecretOP bool
	var yes bool
	cmd := &cobra.Command{
		Use:     "set <KEY=VALUE> [<KEY=VALUE>...]",
		Aliases: []string{"s"},
		Short:   "Define ou atualiza variáveis no vault padrão ou num vault específico",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
			pairs, err := parseKeyValuePairs(args)
			if err != nil {
				return err
			}

			paths := []string{d.Runtime.DefaultEnvPath}
			if setVault != "" {
				vp, perr := appenvs.TargetPath(envPaths(d), setVault)
				if perr != nil {
					return perr
				}
				paths = append(paths, vp)
			}
			asSecret, secretOP, err := appenvs.ResolveSetSecretFlags(
				flagSecret,
				flagSecretOP,
				paths...)
			if err != nil {
				return err
			}

			if secretOP && setVault == "" && !yes {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					return fmt.Errorf(
						"gravar segredos 1Password no vault padrão (env.defaults) pode pedir desbloqueio do 1Password em todo o comando mb; em modo não interativo use --yes para confirmar",
					)
				}
				_ = log.Warn(
					ctx,
					"A gravação da variável no 1Password no vault padrão pode resultar em pedidos frequentes de desbloqueio do 1Password em todo o comando mb.",
				)
				_, _ = fmt.Fprintln(cmd.ErrOrStderr())
				ok, cerr := system.Confirm(
					ctx,
					"Deseja continuar com a gravação da variável no 1Password no vault padrão?",
					cmd.InOrStdin(),
					cmd.ErrOrStderr(),
				)
				if cerr != nil {
					return cerr
				}
				if !ok {
					return fmt.Errorf("cancelado")
				}
			}

			for _, kv := range pairs {
				if err := appenvs.Set(
					d.SecretStore,
					d.OnePassword,
					envPaths(d),
					setVault,
					kv.key,
					kv.value,
					asSecret,
					secretOP,
				); err != nil {
					return err
				}
				if setVault != "" {
					if secretOP {
						_ = log.Info(
							ctx,
							"variável %q guardada no 1Password (vault %q)",
							kv.key,
							setVault,
						)
					} else if asSecret {
						_ = log.Info(
							ctx,
							"variável %q guardada no keyring (vault %q)",
							kv.key,
							setVault,
						)
					} else {
						_ = log.Info(ctx, "variável %q salva no vault %q", kv.key, setVault)
					}
				} else {
					if secretOP {
						_ = log.Info(
							ctx,
							"variável %q guardada no 1Password (vault padrão)",
							kv.key,
						)
					} else if asSecret {
						_ = log.Info(ctx, "variável %q guardada no keyring (vault padrão)", kv.key)
					} else {
						_ = log.Info(ctx, "variável %q salva no vault padrão", kv.key)
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().
		StringVar(&setVault, "vault", "", "Grava no vault informado em vez do vault padrão (ex.: --vault staging)")
	cmd.Flags().
		BoolVar(&flagSecret, "secret", false, "Guarda o valor no keyring do sistema em vez do ficheiro env")
	cmd.Flags().
		BoolVar(&flagSecretOP, "secret-op", false, "Guarda o valor no 1Password (CLI op); a referência op:// fica em ficheiro .opsecrets")
	cmd.Flags().
		BoolVar(&yes, "yes", false, "Confirma gravar com --secret-op no vault padrão sem prompt (útil em CI)")
	cmd.MarkFlagsMutuallyExclusive("secret", "secret-op")
	cmd.GroupID = "commands"
	return cmd
}

type kvPair struct {
	key, value string
}

func parseKeyValuePairs(args []string) ([]kvPair, error) {
	out := make([]kvPair, 0, len(args))
	for _, a := range args {
		k, v, ok := strings.Cut(a, "=")
		if !ok || strings.TrimSpace(k) == "" {
			return nil, fmt.Errorf("esperado KEY=VALOR em cada argumento, obtido %q", a)
		}
		out = append(out, kvPair{key: strings.TrimSpace(k), value: v})
	}
	return out, nil
}
