package alias

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/sliceutil"
	"mb/internal/shared/system"
)

func newUnsetCmd(d deps.Dependencies) *cobra.Command {
	var yes bool
	var mbcliYAML bool
	var vaultStr string

	cmd := &cobra.Command{
		Use:   "unset <nome> [<nome>...]",
		Short: "Remove um ou mais aliases registrados",
		Long: `Remove os aliases guardados com os nomes indicados.

O MB CLI grava de imediato as alterações: os nomes deixam de funcionar em mb run e deixam
de constar nos scripts de shell que o MB gera para o seu perfil (os atalhos deixam de existir
depois de abrir um novo terminal ou fazer source no arquivo de perfil).

A remoção pede confirmação no terminal (uma vez para todo o lote); em CI ou scripts use --yes.

Com --mbcli-yaml, --vault aceita os mesmos rótulos que mb envs list (project, project/<nome>) ou só o nome do submapa.

Se algum nome não existir ou for inválido, o comando termina com erro e nenhum alias é removido.`,
		Example: `  mb alias unset dev
  mb alias unset api worker
  mb alias unset api worker --yes`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())

			names := sliceutil.DedupeStringsPreserveOrder(args)
			for _, name := range names {
				if err := alib.ValidateName(name); err != nil {
					return err
				}
			}

			cfgDir := d.Runtime.ConfigDir
			if mbcliYAML {
				return unsetAliasesMbcli(ctx, cmd, log, names, yes, vaultStr)
			}
			return unsetAliasesConfig(ctx, cmd, log, cfgDir, names, yes, vaultStr)
		},
	}
	cmd.Flags().BoolVar(
		&mbcliYAML, "mbcli-yaml", false,
		"Remove apenas aliases de mbcli.yaml (não altera ~/.config/mb/aliases.yaml nem os scripts de shell)",
	)
	cmd.Flags().StringVar(
		&vaultStr, "vault", "",
		"Vault do slot a remover (vazio = sem vault em ~/.config/mb; com --mbcli-yaml: project, project/<n> ou nome do submapa, como mb envs)",
	)
	cmd.Flags().BoolVarP(
		&yes, "yes", "y", false,
		"Confirma a remoção de todos os aliases listados sem prompt (CI / não interativo)",
	)
	return cmd
}

func unsetAliasesMbcli(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	names []string,
	yes bool,
	vaultStr string,
) error {
	mbcliPath, err := deps.ResolveMbcliYAMLPath()
	if err != nil {
		return err
	}
	vaultInner, err := normalizeMbcliAliasVaultFlag(vaultStr)
	if err != nil {
		return err
	}
	proj, err := deps.ParseMbcliAliases(mbcliPath)
	if err != nil {
		return err
	}
	var missing []string
	for _, name := range names {
		sk := alib.StoreKey(vaultInner, name)
		if _, ok := proj[sk]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"aliases inexistentes em mbcli.yaml: %s (use mb alias unset sem --mbcli-yaml para remover de ~/.config/mb/aliases.yaml)",
			strings.Join(missing, ", "),
		)
	}
	prompt := buildUnsetMbcliBatchPrompt(proj, names, vaultInner)
	if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
		if errors.Is(err, ErrAliasOpCancelled) {
			_ = log.Info(ctx, "Operação cancelada.")
			return nil
		}
		return err
	}
	for _, name := range names {
		if err := deps.RemoveMbcliYAMLAlias(mbcliPath, name, vaultInner); err != nil {
			return err
		}
	}
	if len(names) == 1 {
		_ = log.Info(ctx, "Alias %q removido de mbcli.yaml (%q).", names[0], mbcliPath)
		return nil
	}
	_ = log.Info(
		ctx,
		"Removidos %d aliases de mbcli.yaml (%q): %s.",
		len(names),
		mbcliPath,
		strings.Join(names, ", "),
	)
	return nil
}

func unsetAliasesConfig(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	cfgDir string,
	names []string,
	yes bool,
	vaultStr string,
) error {
	path := alib.FilePath(cfgDir)
	f, err := alib.Load(path)
	if err != nil {
		return err
	}

	var missing []string
	for _, name := range names {
		sk := alib.StoreKey(vaultStr, name)
		if _, ok := f.Aliases[sk]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"aliases inexistentes: %s",
			strings.Join(missing, ", "),
		)
	}

	prompt := buildUnsetBatchPrompt(f, names, vaultStr)
	if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
		if errors.Is(err, ErrAliasOpCancelled) {
			_ = log.Info(ctx, "Operação cancelada.")
			return nil
		}
		return err
	}

	cmdLines := make(map[string]string, len(names))
	for _, name := range names {
		sk := alib.StoreKey(vaultStr, name)
		e := f.Aliases[sk]
		cmdLines[name] = strings.Join(e.Command, " ")
		delete(f.Aliases, sk)
	}
	if err := saveAndRegenerate(cfgDir, f); err != nil {
		return err
	}

	if len(names) == 1 {
		name := names[0]
		_ = log.Info(
			ctx,
			"Alias %q removido (comando era: %s).",
			name,
			cmdLines[name],
		)
		return nil
	}
	_ = log.Info(
		ctx,
		"Removidos %d aliases: %s.",
		len(names),
		strings.Join(names, ", "),
	)
	return nil
}

func buildUnsetMbcliBatchPrompt(
	proj map[string]alib.Entry,
	names []string,
	vaultInner string,
) string {
	if len(names) == 1 {
		return fmt.Sprintf(
			"Deseja remover o alias %q (vault %s) de mbcli.yaml?",
			names[0],
			aliasMbcliLogicalVaultDisplay(vaultInner),
		)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Remover %d aliases de mbcli.yaml?\n\n", len(names))
	for _, name := range names {
		e := proj[alib.StoreKey(vaultInner, name)]
		cmdLine := strings.Join(e.Command, " ")
		fmt.Fprintf(
			&b,
			"- %q: %s\n",
			name,
			truncateForTable(cmdLine, 72),
		)
	}
	b.WriteString("\nConfirmar remoção?")
	return b.String()
}

func buildUnsetBatchPrompt(f *alib.File, names []string, envVault string) string {
	if len(names) == 1 {
		return fmt.Sprintf(
			"Deseja remover o alias %q (vault %s)?",
			names[0],
			configVaultLabel(envVault),
		)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Remover %d aliases?\n\n", len(names))
	for _, name := range names {
		e := f.Aliases[alib.StoreKey(envVault, name)]
		cmdLine := strings.Join(e.Command, " ")
		fmt.Fprintf(
			&b,
			"- %q: %s\n",
			name,
			truncateForTable(cmdLine, 72),
		)
	}
	b.WriteString("\nConfirmar remoção?")
	return b.String()
}
