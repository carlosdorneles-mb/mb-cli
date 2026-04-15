package envs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/deps"
	"mb/internal/shared/sliceutil"
	"mb/internal/shared/system"
	appenvs "mb/internal/usecase/envs"
)

func newUnsetCmd(d deps.Dependencies) *cobra.Command {
	var unsetVault string
	var mbcliYAML bool
	var yes bool
	cmd := &cobra.Command{
		Use:     "unset <KEY> [<KEY>...]",
		Aliases: []string{"u"},
		Short:   "Remove chaves do vault padrão, de um vault nomeado ou do mbcli.yaml (plano, keyring e 1Password)",
		Long: `Remove uma ou mais chaves do mesmo sítio onde mb envs set as gravaria.

Sem --vault, o alvo é o vault padrão (env.defaults e, se existirem, entradas em .secrets / .opsecrets para essa chave). Com --vault <nome>, o alvo é .env.<nome> e os respetivos ficheiros de segredo desse vault. A remoção inclui valor em texto plano, valor no keyring e referência/campo no 1Password quando aplicável. Se a chave não existir, o comando não falha: regista que nada foi removido e segue para as restantes.

Com --mbcli-yaml, remove apenas chaves em envs do mbcli.yaml (raiz de envs ou submapa envs.<nome> com --vault <nome>). Em modo não interativo é obrigatório --yes.`,
		Example: `  # Vault padrão (env.defaults e segredos associados)
  mb envs unset API_URL

  # Vault explícito (.env.<nome>)
  mb envs unset API_URL --vault staging

  # Várias chaves de uma vez
  mb envs unset A B C

  # mbcli.yaml (confirmação automática em CI)
  mb envs unset OLD_FLAG --mbcli-yaml --yes`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvsUnset(cmd, d, envUnsetRunFlags{
				Vault:     unsetVault,
				MbcliYAML: mbcliYAML,
				Yes:       yes,
			}, args)
		},
	}
	cmd.Flags().StringVar(
		&unsetVault,
		"vault",
		"",
		"Remove do vault informado em vez do vault padrão (ex.: --vault staging); com --mbcli-yaml, submapa envs.<nome> no YAML",
	)
	cmd.Flags().BoolVar(
		&mbcliYAML,
		"mbcli-yaml",
		false,
		"Remove apenas chaves em mbcli.yaml (não altera env.defaults nem .env.<vault>)",
	)
	cmd.Flags().BoolVar(
		&yes,
		"yes",
		false,
		"Confirma remoção em mbcli.yaml sem prompt (CI / não interativo)",
	)
	cmd.GroupID = "commands"
	return cmd
}

type envUnsetRunFlags struct {
	Vault     string
	MbcliYAML bool
	Yes       bool
}

func runEnvsUnset(
	cmd *cobra.Command,
	d deps.Dependencies,
	f envUnsetRunFlags,
	args []string,
) error {
	ctx := commandContext(cmd)
	log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
	if f.MbcliYAML {
		return runEnvsUnsetMbcliYAML(ctx, cmd, log, f, args)
	}
	return runEnvsUnsetVaultFiles(ctx, d, log, f.Vault, args)
}

func runEnvsUnsetMbcliYAML(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	f envUnsetRunFlags,
	args []string,
) error {
	keys := sliceutil.DedupeStringsPreserveOrder(args)
	mbcliPath, err := deps.ResolveMbcliYAMLPath()
	if err != nil {
		return err
	}
	missing, err := deps.MbcliYAMLEnvKeysMissing(mbcliPath, f.Vault, keys)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"variáveis inexistentes em mbcli.yaml (vault %s): %s (use mb envs unset sem --mbcli-yaml para remover do diretório de configuração)",
			mbcliUnsetVaultLabel(f.Vault),
			strings.Join(missing, ", "),
		)
	}
	prompt := buildMbcliUnsetEnvPrompt(mbcliPath, f.Vault, keys)
	if !f.Yes {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf(
				"remover variáveis de mbcli.yaml em modo não interativo requer a flag --yes",
			)
		}
		ok, cerr := system.Confirm(ctx, prompt, cmd.InOrStdin(), cmd.ErrOrStderr())
		if cerr != nil {
			return cerr
		}
		if !ok {
			_ = log.Info(ctx, "Operação cancelada.")
			return nil
		}
	}
	if err := deps.RemoveMbcliYAMLEnvKeys(mbcliPath, f.Vault, keys); err != nil {
		return err
	}
	logMbcliUnsetSuccess(ctx, log, mbcliPath, keys)
	return nil
}

func logMbcliUnsetSuccess(
	ctx context.Context,
	log *system.Logger,
	mbcliPath string,
	keys []string,
) {
	if len(keys) == 1 {
		_ = log.Info(
			ctx,
			"Variável %q removida de mbcli.yaml (%q).",
			keys[0],
			mbcliPath,
		)
		return
	}
	_ = log.Info(
		ctx,
		"Removidas %d variáveis de mbcli.yaml (%q): %s.",
		len(keys),
		mbcliPath,
		strings.Join(keys, ", "),
	)
}

func runEnvsUnsetVaultFiles(
	ctx context.Context,
	d deps.Dependencies,
	log *system.Logger,
	unsetVault string,
	args []string,
) error {
	for _, key := range args {
		removed, err := appenvs.Unset(
			d.SecretStore,
			d.OnePassword,
			envPaths(d),
			unsetVault,
			key,
		)
		if err != nil {
			return err
		}
		logEnvUnsetVaultResult(ctx, log, unsetVault, key, removed)
	}
	return nil
}

func logEnvUnsetVaultResult(
	ctx context.Context,
	log *system.Logger,
	vault, key string,
	removed bool,
) {
	if !removed {
		if vault != "" {
			_ = log.Info(ctx, "Não existe variável %q no vault %q", key, vault)
		} else {
			_ = log.Info(ctx, "Não existe variável %q no vault padrão", key)
		}
		return
	}
	if vault != "" {
		_ = log.Info(ctx, "Variável %q removida do vault %q", key, vault)
	} else {
		_ = log.Info(ctx, "Variável %q removida do vault padrão", key)
	}
}

func mbcliUnsetVaultLabel(vault string) string {
	if vault == "" {
		return "raiz (project)"
	}
	return vault
}

func buildMbcliUnsetEnvPrompt(mbcliPath, vault string, keys []string) string {
	def, byV, _ := deps.ParseMbcliProjectEnvs(mbcliPath)
	if len(keys) == 1 {
		k := keys[0]
		var v string
		if vault == "" {
			v = def[k]
		} else if inner, ok := byV[vault]; ok {
			v = inner[k]
		}
		return fmt.Sprintf("Deseja remover a variável %q (= %q) de mbcli.yaml?", k, v)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Remover %d variáveis de mbcli.yaml?\n\n", len(keys))
	for _, k := range keys {
		var v string
		if vault == "" {
			v = def[k]
		} else if inner, ok := byV[vault]; ok {
			v = inner[k]
		}
		fmt.Fprintf(&b, "- %q = %q\n", k, v)
	}
	b.WriteString("\nConfirmar remoção?")
	return b.String()
}
