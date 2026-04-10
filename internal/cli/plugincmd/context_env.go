package plugincmd

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"mb/internal/infra/sqlite"
)

// appendPluginInvocationEnv adds MB_CTX_* variables so plugin scripts can inspect how mb was invoked.
// argv is typically os.Args; plugin is the manifest plugin (not a synthetic flag entrypoint copy).
// changedPluginFlagNames are long flag names (Cobra/pflag), sorted for stability.
func appendPluginInvocationEnv(
	merged []string,
	cmd *cobra.Command,
	plugin sqlite.Plugin,
	argv []string,
	configDir string,
	changedPluginFlagNames []string,
) []string {
	inv := strings.Join(argv, " ")
	merged = append(merged, "MB_CTX_INVOCATION="+inv)
	if configDir != "" {
		merged = append(merged, "MB_CTX_CONFIG_DIR="+configDir)
	}

	cp := strings.TrimSpace(plugin.CommandPath)
	merged = append(merged, "MB_CTX_COMMAND_PATH="+cp)

	name := commandNameFromPath(cp)
	merged = append(merged, "MB_CTX_COMMAND_NAME="+name)

	parent := parentCommandPath(cp)
	merged = append(merged, "MB_CTX_PARENT_COMMAND_PATH="+parent)

	if cmd != nil {
		merged = append(merged, "MB_CTX_COBR_COMMAND_PATH="+cmd.CommandPath())
	}

	merged = append(merged, "MB_CTX_PLUGIN_FLAGS="+strings.Join(changedPluginFlagNames, " "))

	merged = append(merged, "MB_CTX_PEER_COMMANDS="+peerCommandsJSON(cmd))
	merged = append(merged, "MB_CTX_CHILD_COMMANDS="+visibleChildCommandsJSON(cmd))
	merged = append(merged, "MB_CTX_HIDDEN_CHILD_COMMANDS="+hiddenChildCommandsJSON(cmd))
	merged = append(merged, "MB_CTX_CHILD_COMMAND_ALIASES="+childCommandAliasesJSON(cmd))

	return merged
}

func commandNameFromPath(commandPath string) string {
	commandPath = strings.TrimSpace(commandPath)
	if commandPath == "" {
		return ""
	}
	idx := strings.LastIndex(commandPath, "/")
	if idx < 0 || idx == len(commandPath)-1 {
		return commandPath
	}
	return commandPath[idx+1:]
}

func parentCommandPath(commandPath string) string {
	commandPath = strings.TrimSpace(commandPath)
	if commandPath == "" {
		return ""
	}
	idx := strings.LastIndex(commandPath, "/")
	if idx <= 0 {
		return ""
	}
	return commandPath[:idx]
}

func peerCommandsJSON(cmd *cobra.Command) string {
	if cmd == nil {
		return "[]"
	}
	parent := cmd.Parent()
	if parent == nil {
		return "[]"
	}
	self := cmd.Name()
	var peers []string
	for _, c := range parent.Commands() {
		if c.Hidden {
			continue
		}
		if c.Name() == self {
			continue
		}
		peers = append(peers, c.Name())
	}
	sort.Strings(peers)
	b, err := json.Marshal(peers)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// childAliasEntry is serialized into MB_CTX_CHILD_COMMAND_ALIASES.
type childAliasEntry struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

func visibleChildCommandsJSON(cmd *cobra.Command) string {
	if cmd == nil {
		return "[]"
	}
	names := make([]string, 0)
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		names = append(names, c.Name())
	}
	sort.Strings(names)
	b, err := json.Marshal(names)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func hiddenChildCommandsJSON(cmd *cobra.Command) string {
	if cmd == nil {
		return "[]"
	}
	names := make([]string, 0)
	for _, c := range cmd.Commands() {
		if !c.Hidden {
			continue
		}
		names = append(names, c.Name())
	}
	sort.Strings(names)
	b, err := json.Marshal(names)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func childCommandAliasesJSON(cmd *cobra.Command) string {
	if cmd == nil {
		return "[]"
	}
	entries := make([]childAliasEntry, 0)
	for _, c := range cmd.Commands() {
		if len(c.Aliases) == 0 {
			continue
		}
		aliases := append([]string(nil), c.Aliases...)
		sort.Strings(aliases)
		entries = append(entries, childAliasEntry{Name: c.Name(), Aliases: aliases})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	b, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// changedLocalPluginFlags lists changed flags on the command's local FlagSet, excluding readme.
func changedLocalPluginFlags(cmd *cobra.Command) []string {
	if cmd == nil {
		return nil
	}
	var names []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name == "readme" {
			return
		}
		names = append(names, f.Name)
	})
	sort.Strings(names)
	return names
}
