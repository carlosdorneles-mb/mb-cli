package update

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/spf13/cobra"

	"mb/internal/shared/system"
)

// findToolsUpdateAllCmd returns the root-level "tools" subcommand if it defines the update-all flag
// (from plugin manifest).
func findToolsUpdateAllCmd(root *cobra.Command) *cobra.Command {
	if root == nil {
		return nil
	}
	for _, c := range root.Commands() {
		if c.Name() != "tools" {
			continue
		}
		if c.Flags().Lookup("update-all") == nil {
			continue
		}
		return c
	}
	return nil
}

// RunToolsUpdateAllPhase runs "mb tools --update-all" via a nested root.Execute when the tools
// plugin exposes --update-all. If toolsOnlyExclusive is true (--only-tools with no other --only-*),
// logs a warning when the phase is skipped.
func RunToolsUpdateAllPhase(
	ctx context.Context,
	cmd *cobra.Command,
	log *system.Logger,
	toolsOnlyExclusive bool,
) error {
	root := cmd.Root()
	if findToolsUpdateAllCmd(root) == nil {
		if toolsOnlyExclusive {
			_ = log.Warn(
				ctx,
				"Comando tools ou flag --update-all indisponível; fase tools ignorada.",
			)
		}
		return nil
	}

	_ = log.Info(ctx, "Atualizando ferramentas (mb tools --update-all)...")

	prev, wasSet := saveRootArgs(root)
	root.SetArgs([]string{"tools", "--update-all"})
	err := root.ExecuteContext(ctx)
	restoreRootArgs(root, prev, wasSet)
	return err
}

func saveRootArgs(root *cobra.Command) (args []string, wasSet bool) {
	if root == nil {
		return nil, false
	}
	rv := reflect.ValueOf(root).Elem()
	fa := rv.FieldByName("args")
	if !fa.IsValid() {
		return nil, false
	}
	sl := reflect.NewAt(fa.Type(), unsafe.Pointer(fa.UnsafeAddr())).Elem()
	if sl.IsNil() {
		return nil, false
	}
	out := make([]string, sl.Len())
	for i := range out {
		out[i] = sl.Index(i).String()
	}
	return out, true
}

func restoreRootArgs(root *cobra.Command, args []string, wasSet bool) {
	if root == nil {
		return
	}
	if !wasSet {
		root.SetArgs(nil)
		return
	}
	root.SetArgs(append([]string(nil), args...))
}
