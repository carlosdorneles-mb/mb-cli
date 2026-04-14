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
		Use:     "set <KEY[=VALOR]> [<KEY[=VALOR]>...]",
		Aliases: []string{"s"},
		Short:   "Define ou atualiza variáveis no vault padrão ou num vault específico",
		Long: "Com --secret ou --secret-op, pode omitir o valor: use só a chave (ex.: API_KEY) e o MB pede o valor " +
			"com gum input --password (um prompt por chave), sem gravar o segredo no histórico da shell. " +
			"Com valor na linha de comandos (KEY=VALOR), não há prompt.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

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
			secretMode := asSecret || secretOP

			pairs, err := parseEnvSetArgs(args, secretMode)
			if err != nil {
				return err
			}

			if envSetAnyNeedsPrompt(pairs) && !term.IsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf(
					"definir chave sem valor (sem '=') com --secret ou --secret-op requer um terminal interativo; " +
						"use KEY=VALOR na linha de comandos ou execute num TTY",
				)
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

			for i := range pairs {
				if !pairs[i].needsPrompt {
					continue
				}
				val, perr := system.PromptSecretValue(ctx, pairs[i].key)
				if perr != nil {
					return perr
				}
				if strings.TrimSpace(val) == "" {
					return fmt.Errorf("o valor para %q não pode ser vazio", pairs[i].key)
				}
				pairs[i].value = val
			}

			if secretMode && envSetAnyInlineSecretValue(pairs) {
				_ = log.Warn(
					ctx,
					"Passar o segredo na linha de comandos (KEY=VALOR) não é seguro: o valor pode ficar no histórico da shell "+
						"e em registos de processos. Prefira `mb envs set CHAVE --secret` ou `mb envs set CHAVE --secret-op` sem '=' "+
						"para o valor ser pedido com mascaramento.",
				)
				_, _ = fmt.Fprintln(cmd.ErrOrStderr())
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
		BoolVar(&flagSecret, "secret", false, "Guarda o valor no keyring; com KEY sem '=' pede o valor com gum input --password (um por chave)")
	cmd.Flags().
		BoolVar(&flagSecretOP, "secret-op", false, "Guarda no 1Password (op); com KEY sem '=' pede o valor com gum input --password (um por chave)")
	cmd.Flags().
		BoolVar(&yes, "yes", false, "Confirma gravar com --secret-op no vault padrão sem prompt (útil em CI)")
	cmd.MarkFlagsMutuallyExclusive("secret", "secret-op")
	cmd.GroupID = "commands"
	return cmd
}

type kvPair struct {
	key, value  string
	needsPrompt bool
}

func envSetAnyNeedsPrompt(pairs []kvPair) bool {
	for _, p := range pairs {
		if p.needsPrompt {
			return true
		}
	}
	return false
}

// envSetAnyInlineSecretValue reports whether any pair used KEY=VALOR on the command line (not prompted).
func envSetAnyInlineSecretValue(pairs []kvPair) bool {
	for _, p := range pairs {
		if !p.needsPrompt {
			return true
		}
	}
	return false
}

// parseEnvSetArgs parses positional arguments for mb envs set.
// When secretMode is false, every argument must be KEY=VALOR.
// When secretMode is true, an argument without '=' is treated as KEY with the value to be prompted later.
func parseEnvSetArgs(args []string, secretMode bool) ([]kvPair, error) {
	out := make([]kvPair, 0, len(args))
	for _, a := range args {
		k, v, ok := strings.Cut(a, "=")
		if !ok {
			if !secretMode {
				return nil, fmt.Errorf("esperado KEY=VALOR em cada argumento, obtido %q", a)
			}
			key := strings.TrimSpace(a)
			if key == "" {
				return nil, fmt.Errorf("chave vazia em argumento %q", a)
			}
			out = append(out, kvPair{key: key, needsPrompt: true})
			continue
		}
		key := strings.TrimSpace(k)
		if key == "" {
			return nil, fmt.Errorf("esperado KEY=VALOR em cada argumento, obtido %q", a)
		}
		out = append(out, kvPair{key: key, value: v, needsPrompt: false})
	}
	return out, nil
}
