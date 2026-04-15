package alias

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

func newSetCmd(d deps.Dependencies) *cobra.Command {
	var vaultStr string
	var yes bool
	var mbcliYAML bool

	cmd := &cobra.Command{
		Use:   "set <name> [--vault <vault>] [-- <command>...]",
		Short: "Define ou atualiza um alias",
		Long: `Associa um nome curto a um comando (com argumentos). Os dados ficam na configuração
do MB CLI e são refletidos nos scripts de shell gerados automaticamente.

Com --vault, apenas mb run <nome> usa esse vault de ambiente injetando as variáveis do vault ao comando executado.
Invocar o mesmo nome diretamente no shell, executa o comando com o ambiente normal do terminal sem a injeção
de variáveis de ambiente do vault.

Para alterar só o vault de um alias já existente, use --vault (e opcionalmente valor vazio
para desassociar o vault) com um único argumento: o nome do alias.

Com --mbcli-yaml, --vault aceita os mesmos rótulos que mb envs list (project, project/<nome>) ou só o nome do submapa.

Alterar ou remover um alias existente pede confirmação no terminal; em CI ou scripts use --yes.`,
		Example: `
  # Define um alias para o comando docker compose up
  mb alias set dev -- docker compose up

  # Define um alias para o comando npm run build com o vault staging
  mb alias set dev --vault staging -- npm run build

  # Remove o vault associado ao alias
  mb alias set dev --vault ""

  # Atualiza o alias associado ao vault staging
  mb alias set dev --vault staging

  # Define um alias com valor literal para uma variável de ambiente
  mb alias set api --vault staging -- sh -c 'echo API $KEY'

  # Alias com comando em várias linhas (use \ no fim da linha antes de Enter no shell)
  mb alias set dev -- \
    docker compose \
    up -d
  
  # Sem prompt (automação); o comando também pode continuar em várias linhas
  mb alias set dev --yes -- \
    docker compose up -d`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultFlag := cmd.Flags().Lookup("vault")
			vaultChanged := vaultFlag != nil && vaultFlag.Changed
			return runAliasSet(cmd, d, aliasSetRunInput{
				Vault:            vaultStr,
				VaultFlagChanged: vaultChanged,
				Yes:              yes,
				MbcliYAML:        mbcliYAML,
			}, args)
		},
	}

	cmd.Flags().StringVar(
		&vaultStr, "vault", "",
		"Vault de ambiente usado apenas com mb run; vazio remove a associação ao vault (com --mbcli-yaml: project, project/<n> ou nome do submapa, alinhado a mb envs)",
	)
	cmd.Flags().BoolVar(
		&mbcliYAML, "mbcli-yaml", false,
		"Grava ou atualiza o alias em mbcli.yaml (resolvido por MBCLI_YAML_PATH / MBCLI_PROJECT_ROOT); não regenera os scripts de shell",
	)
	cmd.Flags().BoolVarP(
		&yes, "yes", "y", false,
		"Confirma alterações a aliases existentes sem prompt (CI / não interativo)",
	)
	return cmd
}

type aliasSetRunInput struct {
	Vault            string
	VaultFlagChanged bool
	Yes              bool
	MbcliYAML        bool
}

func commandContext(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func runAliasSet(
	cmd *cobra.Command,
	d deps.Dependencies,
	in aliasSetRunInput,
	args []string,
) error {
	ctx := commandContext(cmd)
	log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

	name := args[0]
	if err := alib.ValidateName(name); err != nil {
		return err
	}

	cfgDir := d.Runtime.ConfigDir
	path := alib.FilePath(cfgDir)
	f, err := alib.Load(path)
	if err != nil {
		return err
	}

	if in.MbcliYAML {
		mbcliPath, err := deps.ResolveMbcliYAMLPath()
		if err != nil {
			return err
		}
		return runSetMbcliYAML(
			ctx, cmd, log, mbcliPath, name, args, in.VaultFlagChanged, in.Vault, in.Yes,
		)
	}

	if len(args) == 1 {
		if !in.VaultFlagChanged {
			return errors.New(
				"indique --vault para atualizar só o vault, ou o comando após o nome " +
					"(ex.: mb alias set <nome> -- <cmd>)",
			)
		}
		return runAliasSetVaultOnly(ctx, cmd, log, cfgDir, f, name, in.Vault, in.Yes)
	}

	return runAliasSetWithCommand(ctx, cmd, log, cfgDir, f, name, args, in)
}

func runAliasSetVaultOnly(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	cfgDir string,
	f *alib.File,
	name string,
	vaultStr string,
	yes bool,
) error {
	oldKey, e, n := alib.FindUniqueStoreKeyForDisplayName(f.Aliases, name)
	if n == 0 {
		hint := "defina o comando primeiro (ex.: mb alias set %s -- <cmd>)"
		return fmt.Errorf("alias %q ainda não existe; "+hint, name, name)
	}
	if n > 1 {
		return fmt.Errorf(
			"há vários aliases com o nome %q; remova os que não precisa ou use "+
				"mb alias set %s --vault <vault-do-slot> -- <cmd> para escolher o slot (nome+vault)",
			name, name,
		)
	}
	oldV := e.EnvVault
	prompt := fmt.Sprintf(
		"Deseja alterar o vault do alias %q para %s?",
		name,
		configVaultLabel(vaultStr),
	)
	if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
		if errors.Is(err, ErrAliasOpCancelled) {
			_ = log.Info(ctx, "Operação cancelada.")
			return nil
		}
		return err
	}
	delete(f.Aliases, oldKey)
	e.EnvVault = vaultStr
	if err := alib.ValidateEntry(e); err != nil {
		return err
	}
	f.Aliases[alib.StoreKey(e.EnvVault, name)] = e
	if err := saveAndRegenerate(cfgDir, f); err != nil {
		return err
	}
	feedbackVaultOnly(ctx, log, name, oldV, vaultStr, false)
	return nil
}

func runAliasSetWithCommand(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	cfgDir string,
	f *alib.File,
	name string,
	args []string,
	in aliasSetRunInput,
) error {
	cmdArgv := append([]string(nil), args[1:]...)
	slotVault := ""
	if in.VaultFlagChanged {
		slotVault = in.Vault
	}
	sk := alib.StoreKey(slotVault, name)
	e, had := f.Aliases[sk]
	newE := alib.Entry{Command: cmdArgv}
	if in.VaultFlagChanged {
		newE.EnvVault = in.Vault
	} else if had {
		newE.EnvVault = e.EnvVault
	}
	if err := alib.ValidateEntry(newE); err != nil {
		return err
	}

	if had {
		if err := confirmExistingAliasChange(
			ctx,
			cmd,
			in.Yes,
			name,
			e,
			newE,
			false,
		); err != nil {
			if errors.Is(err, ErrAliasOpCancelled) {
				_ = log.Info(ctx, "Operação cancelada.")
				return nil
			}
			return err
		}
	}

	f.Aliases[sk] = newE
	if err := saveAndRegenerate(cfgDir, f); err != nil {
		return err
	}

	if !had {
		feedbackNewAlias(ctx, log, name, newE, false)
		return nil
	}
	feedbackUpdatedAlias(ctx, log, name, e, newE, false)
	return nil
}
