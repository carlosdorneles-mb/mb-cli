package alias

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

func dedupeNamesPreserveOrder(names []string) []string {
	seen := make(map[string]bool, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		if seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return out
}

func newUnsetCmd(d deps.Dependencies) *cobra.Command {
	var yes bool
	var mbcliYAML bool

	cmd := &cobra.Command{
		Use:   "unset <nome> [<nome>...]",
		Short: "Remove um ou mais aliases registrados",
		Long: `Remove os aliases guardados com os nomes indicados.

O MB CLI grava de imediato as alterações: os nomes deixam de funcionar em mb run e deixam
de constar nos scripts de shell que o MB gera para o seu perfil (os atalhos deixam de existir
depois de abrir um novo terminal ou fazer source no arquivo de perfil).

A remoção pede confirmação no terminal (uma vez para todo o lote); em CI ou scripts use --yes.

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

			names := dedupeNamesPreserveOrder(args)
			for _, name := range names {
				if err := alib.ValidateName(name); err != nil {
					return err
				}
			}

			cfgDir := d.Runtime.ConfigDir
			if mbcliYAML {
				mbcliPath, err := deps.ResolveMbcliYAMLPath()
				if err != nil {
					return err
				}
				proj, err := deps.ParseMbcliAliases(mbcliPath)
				if err != nil {
					return err
				}
				var missing []string
				for _, name := range names {
					if _, ok := proj[name]; !ok {
						missing = append(missing, name)
					}
				}
				if len(missing) > 0 {
					return fmt.Errorf(
						"aliases inexistentes em mbcli.yaml: %s (use mb alias unset sem --mbcli-yaml para remover de ~/.config/mb/aliases.yaml)",
						strings.Join(missing, ", "),
					)
				}
				prompt := buildUnsetMbcliBatchPrompt(proj, names)
				if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
					if errors.Is(err, ErrAliasOpCancelled) {
						_ = log.Info(ctx, "Operação cancelada.")
						return nil
					}
					return err
				}
				for _, name := range names {
					if err := deps.RemoveMbcliYAMLAlias(mbcliPath, name); err != nil {
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

			path := alib.FilePath(cfgDir)
			f, err := alib.Load(path)
			if err != nil {
				return err
			}

			var missing []string
			for _, name := range names {
				if _, ok := f.Aliases[name]; !ok {
					missing = append(missing, name)
				}
			}
			if len(missing) > 0 {
				return fmt.Errorf(
					"aliases inexistentes: %s",
					strings.Join(missing, ", "),
				)
			}

			prompt := buildUnsetBatchPrompt(f, names)
			if err := requireConfirmOrYes(ctx, cmd, yes, prompt); err != nil {
				if errors.Is(err, ErrAliasOpCancelled) {
					_ = log.Info(ctx, "Operação cancelada.")
					return nil
				}
				return err
			}

			cmdLines := make(map[string]string, len(names))
			for _, name := range names {
				e := f.Aliases[name]
				cmdLines[name] = strings.Join(e.Command, " ")
				delete(f.Aliases, name)
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
		},
	}
	cmd.Flags().BoolVar(
		&mbcliYAML, "mbcli-yaml", false,
		"Remove apenas aliases de mbcli.yaml (não altera ~/.config/mb/aliases.yaml nem os scripts de shell)",
	)
	cmd.Flags().BoolVarP(
		&yes, "yes", "y", false,
		"Confirma a remoção de todos os aliases listados sem prompt (CI / não interativo)",
	)
	return cmd
}

func buildUnsetMbcliBatchPrompt(proj map[string]alib.Entry, names []string) string {
	if len(names) == 1 {
		return fmt.Sprintf("Deseja remover o alias %q de mbcli.yaml?", names[0])
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Remover %d aliases de mbcli.yaml?\n\n", len(names))
	for _, name := range names {
		e := proj[name]
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

func buildUnsetBatchPrompt(f *alib.File, names []string) string {
	if len(names) == 1 {
		return fmt.Sprintf("Deseja remover o alias %q?", names[0])
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Remover %d aliases?\n\n", len(names))
	for _, name := range names {
		e := f.Aliases[name]
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
