package plugincmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"mb/internal/shared/system"
)

// persistentShorthandSet returns each single-letter shorthand declared on root persistent flags.
// Call after root has registered PersistentFlags (see internal/cli/root NewRootCmd before Attach).
func persistentShorthandSet(root *cobra.Command) map[string]struct{} {
	out := make(map[string]struct{})
	if root == nil {
		return out
	}
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		s := f.Shorthand
		if s != "" && len([]rune(s)) == 1 {
			out[s] = struct{}{}
		}
	})
	return out
}

func globalShorthandReserved(m map[string]struct{}, sh string) bool {
	if m == nil {
		return false
	}
	_, ok := m[sh]
	return ok
}

func pluginFlagShortDisabledDebug(
	dbgLog *system.Logger,
	commandPath, flagName, shorthand, reason string,
) {
	if dbgLog == nil {
		return
	}
	_ = dbgLog.Debug(
		context.Background(),
		"plugin %s: shorthand %q do flag %q desativado (%s)",
		commandPath,
		shorthand,
		flagName,
		reason,
	)
}

// registerReadmeFlag registers --readme, preferring -r unless r is reserved globally or already used in usedShorts.
func registerReadmeFlag(
	cmd *cobra.Command,
	commandPath, readmePath string,
	dbgLog *system.Logger,
	usedShorts map[string]bool,
	globalShorts map[string]struct{},
) {
	if readmePath == "" {
		return
	}
	const readmeShort = "r"
	resGlob := globalShorthandReserved(globalShorts, readmeShort)
	used := usedShorts != nil && usedShorts[readmeShort]
	if resGlob || used {
		cmd.Flags().Bool("readme", false, readmeFlagDesc)
		switch {
		case used:
			pluginFlagShortDisabledDebug(
				dbgLog,
				commandPath,
				"readme",
				readmeShort,
				"já usado neste comando",
			)
		default:
			pluginFlagShortDisabledDebug(
				dbgLog,
				commandPath,
				"readme",
				readmeShort,
				"reservado pela CLI mb",
			)
		}
		return
	}
	cmd.Flags().BoolP("readme", readmeShort, false, readmeFlagDesc)
	if usedShorts != nil {
		usedShorts[readmeShort] = true
	}
}
