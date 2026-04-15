package envs

import (
	"context"
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
	var mbcliYAML bool
	cmd := &cobra.Command{
		Use:     "set <KEY[=VALOR]> [<KEY[=VALOR]>...]",
		Aliases: []string{"s"},
		Short: "Define ou atualiza variáveis persistidas: texto plano, keyring (--secret) ou 1Password (--secret-op)",
		Long: `Grava variáveis de ambiente persistentes usadas pelo MB (plugins, mb run, etc.).

Sem --secret nem --secret-op, cada argumento tem de ser CHAVE=VALOR. O valor fica em env.defaults (vault padrão) ou em .env.<nome> quando usa --vault <nome>.

Com --secret, o valor vai para o keyring e a chave fica registada em .secrets. Com --secret-op, o valor vai para o 1Password e a referência op:// fica em .opsecrets. Com uma destas flags pode passar só a chave (sem '=') para o MB pedir o valor com gum input --password (um prompt por chave), evitando deixar o segredo no histórico da shell. Com CHAVE=VALOR na linha de comandos não há prompt.

A preferência keyring/op pode também vir da variável MB_ENVS_SECRET_STORE ou dos ficheiros de ambiente já considerados pelo MB.

Com --mbcli-yaml, grava na chave envs do mbcli.yaml do repositório (apenas texto plano; não pode combinar com --secret nem --secret-op). Alterar valores já existentes pode pedir confirmação; em CI use --yes.

Gravar com --secret-op no vault padrão pode implicar confirmação extra; em modo não interativo use --yes.`,
		Example: `  # Texto plano no vault padrão
  mb envs set API_URL=https://api.example.com

  # Vários pares
  mb envs set LOG_LEVEL=debug FEATURE=on

  # Vault nomeado (.env.<nome>)
  mb envs set NODE_ENV=production --vault staging

  # Keyring: pede o valor de forma mascarada
  mb envs set API_TOKEN --secret

  # 1Password no vault prod
  mb envs set DB_PASSWORD --secret-op --vault prod

  # Variáveis partilhadas no repositório (mbcli.yaml)
  mb envs set CI=true --mbcli-yaml`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

			if mbcliYAML {
				if flagSecret || flagSecretOP {
					return fmt.Errorf(
						"--mbcli-yaml não pode ser usado com --secret nem --secret-op (variáveis em mbcli.yaml são só em texto)",
					)
				}
				mbcliPath, err := deps.ResolveMbcliYAMLPath()
				if err != nil {
					return err
				}
				pairs, err := parseEnvSetArgs(args, false)
				if err != nil {
					return err
				}
				pairMap := make(map[string]string, len(pairs))
				for _, kv := range pairs {
					pairMap[kv.key] = kv.value
				}
				def, byV, err := deps.ParseMbcliProjectEnvs(mbcliPath)
				if err != nil {
					return err
				}
				var changeLines []string
				for _, kv := range pairs {
					var old string
					found := false
					if setVault == "" {
						old, found = def[kv.key]
					} else {
						if inner, ok := byV[setVault]; ok {
							old, found = inner[kv.key]
						}
					}
					if found && old != kv.value {
						changeLines = append(
							changeLines,
							fmt.Sprintf("- %s: %q -> %q", kv.key, old, kv.value),
						)
					}
				}
				if len(changeLines) > 0 && !yes {
					if !term.IsTerminal(int(os.Stdin.Fd())) {
						return fmt.Errorf(
							"atualizar variáveis existentes em mbcli.yaml requer terminal interativo ou use a flag --yes",
						)
					}
					prompt := fmt.Sprintf(
						"Atualizar variáveis no mbcli.yaml?\n\n%s\n\nConfirmar?",
						strings.Join(changeLines, "\n"),
					)
					ok, cerr := system.Confirm(ctx, prompt, cmd.InOrStdin(), cmd.ErrOrStderr())
					if cerr != nil {
						return cerr
					}
					if !ok {
						return fmt.Errorf("cancelado")
					}
				}
				if err := deps.UpsertMbcliYAMLEnvs(mbcliPath, setVault, pairMap); err != nil {
					return err
				}
				for _, kv := range pairs {
					_ = log.Info(ctx, "Variável %q salva em mbcli.yaml (%q).", kv.key, mbcliPath)
				}
				return nil
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
		BoolVar(&yes, "yes", false,
			"Confirma sem prompt: --secret-op no vault padrão ou alterações em mbcli.yaml com --mbcli-yaml (útil em CI)",
		)
	cmd.Flags().BoolVar(
		&mbcliYAML,
		"mbcli-yaml",
		false,
		"Grava no mbcli.yaml do repositório (chave envs) em vez de env.defaults ou .env.<vault>",
	)
	cmd.MarkFlagsMutuallyExclusive("secret", "secret-op")
	cmd.MarkFlagsMutuallyExclusive("mbcli-yaml", "secret", "secret-op")
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
