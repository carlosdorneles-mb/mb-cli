package update

import (
	"context"

	"github.com/spf13/cobra"

	"mb/internal/shared/system"
)

// findMachineUpdateCmd reports whether the root has plugin command mb machine update registered.
func findMachineUpdateCmd(root *cobra.Command) bool {
	if root == nil {
		return false
	}
	for _, c := range root.Commands() {
		if c.Name() != "machine" {
			continue
		}
		for _, sub := range c.Commands() {
			if sub.Name() == "update" {
				return true
			}
		}
	}
	return false
}

// RunMachineUpdatePhase runs "mb machine update" via nested root.Execute when the plugin is registered.
// If systemOnlyExclusive is true (--only-system with no other --only-*), logs a warning when skipped.
func RunMachineUpdatePhase(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	systemOnlyExclusive bool,
) error {
	root := cmd.Root()
	if !findMachineUpdateCmd(root) {
		if systemOnlyExclusive {
			_ = log.Warn(
				ctx,
				"Comando machine update indisponível; fase de sistema ignorada (instale o plugin machine e execute mb plugins sync).",
			)
		}
		return nil
	}

	_ = log.Info(ctx, "Atualizando pacotes do sistema (mb machine update)...")

	prev, wasSet := saveRootArgs(root)
	root.SetArgs([]string{"machine", "update"})
	err := root.ExecuteContext(ctx)
	restoreRootArgs(root, prev, wasSet)
	return err
}
