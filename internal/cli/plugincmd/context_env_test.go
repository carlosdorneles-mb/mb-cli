package plugincmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/infra/sqlite"
)

func TestCommandNameFromPath(t *testing.T) {
	t.Parallel()
	if got, want := commandNameFromPath("tools/vscode"), "vscode"; got != want {
		t.Errorf("commandNameFromPath(tools/vscode) = %q, want %q", got, want)
	}
	if got, want := commandNameFromPath("hello"), "hello"; got != want {
		t.Errorf("commandNameFromPath(hello) = %q, want %q", got, want)
	}
	if got, want := commandNameFromPath(""), ""; got != want {
		t.Errorf("commandNameFromPath(empty) = %q, want %q", got, want)
	}
}

func TestParentCommandPath(t *testing.T) {
	t.Parallel()
	if got, want := parentCommandPath("tools/vscode"), "tools"; got != want {
		t.Errorf("parentCommandPath = %q, want %q", got, want)
	}
	if got, want := parentCommandPath("hello"), ""; got != want {
		t.Errorf("parentCommandPath(hello) = %q, want %q", got, want)
	}
}

func TestAppendPluginInvocationEnv(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "mb"}
	tools := &cobra.Command{Use: "tools", Hidden: false}
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}
	tools.AddCommand(a)
	tools.AddCommand(b)
	root.AddCommand(tools)

	plugin := sqlite.Plugin{CommandPath: "tools/a"}
	merged := appendPluginInvocationEnv(
		nil,
		a,
		plugin,
		[]string{"/bin/mb", "tools", "a", "-i"},
		"/cfg",
		[]string{"install"},
	)

	m := map[string]string{}
	for _, e := range merged {
		i := strings.IndexByte(e, '=')
		if i <= 0 {
			t.Fatalf("bad env entry: %q", e)
		}
		m[e[:i]] = e[i+1:]
	}

	if m["MB_CTX_INVOCATION"] != "/bin/mb tools a -i" {
		t.Errorf("MB_CTX_INVOCATION = %q", m["MB_CTX_INVOCATION"])
	}
	if m["MB_CTX_CONFIG_DIR"] != "/cfg" {
		t.Errorf("MB_CTX_CONFIG_DIR = %q", m["MB_CTX_CONFIG_DIR"])
	}
	if m["MB_CTX_COMMAND_PATH"] != "tools/a" {
		t.Errorf("MB_CTX_COMMAND_PATH = %q", m["MB_CTX_COMMAND_PATH"])
	}
	if m["MB_CTX_COMMAND_NAME"] != "a" {
		t.Errorf("MB_CTX_COMMAND_NAME = %q", m["MB_CTX_COMMAND_NAME"])
	}
	if m["MB_CTX_PARENT_COMMAND_PATH"] != "tools" {
		t.Errorf("MB_CTX_PARENT_COMMAND_PATH = %q", m["MB_CTX_PARENT_COMMAND_PATH"])
	}
	if m["MB_CTX_COBR_COMMAND_PATH"] != "mb tools a" {
		t.Errorf("MB_CTX_COBR_COMMAND_PATH = %q", m["MB_CTX_COBR_COMMAND_PATH"])
	}
	if m["MB_CTX_PLUGIN_FLAGS"] != "install" {
		t.Errorf("MB_CTX_PLUGIN_FLAGS = %q", m["MB_CTX_PLUGIN_FLAGS"])
	}

	var peers []string
	if err := json.Unmarshal([]byte(m["MB_CTX_PEER_COMMANDS"]), &peers); err != nil {
		t.Fatalf("MB_CTX_PEER_COMMANDS json: %v", err)
	}
	if len(peers) != 1 || peers[0] != "b" {
		t.Errorf("MB_CTX_PEER_COMMANDS = %v, want [b]", peers)
	}

	var child, hid []string
	if err := json.Unmarshal([]byte(m["MB_CTX_CHILD_COMMANDS"]), &child); err != nil {
		t.Fatalf("MB_CTX_CHILD_COMMANDS json: %v", err)
	}
	if len(child) != 0 {
		t.Errorf("MB_CTX_CHILD_COMMANDS = %v, want []", child)
	}
	if err := json.Unmarshal([]byte(m["MB_CTX_HIDDEN_CHILD_COMMANDS"]), &hid); err != nil {
		t.Fatalf("MB_CTX_HIDDEN_CHILD_COMMANDS json: %v", err)
	}
	if len(hid) != 0 {
		t.Errorf("MB_CTX_HIDDEN_CHILD_COMMANDS = %v, want []", hid)
	}
	if m["MB_CTX_CHILD_COMMAND_ALIASES"] != "[]" {
		t.Errorf("MB_CTX_CHILD_COMMAND_ALIASES = %q, want []", m["MB_CTX_CHILD_COMMAND_ALIASES"])
	}
}

func TestPeerCommandsJSON_nilCmd(t *testing.T) {
	t.Parallel()
	if got := peerCommandsJSON(nil); got != "[]" {
		t.Errorf("peerCommandsJSON(nil) = %q", got)
	}
}

func TestVisibleHiddenChildCommandsJSON_nilCmd(t *testing.T) {
	t.Parallel()
	if got := visibleChildCommandsJSON(nil); got != "[]" {
		t.Errorf("visibleChildCommandsJSON(nil) = %q", got)
	}
	if got := hiddenChildCommandsJSON(nil); got != "[]" {
		t.Errorf("hiddenChildCommandsJSON(nil) = %q", got)
	}
	if got := childCommandAliasesJSON(nil); got != "[]" {
		t.Errorf("childCommandAliasesJSON(nil) = %q", got)
	}
}

func TestChildCommandsEnv_partitionAndAliases(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "mb"}
	tools := &cobra.Command{Use: "tools", Hidden: false}
	leafA := &cobra.Command{Use: "a"}
	leafB := &cobra.Command{Use: "b", Aliases: []string{"bee", "beta"}}
	hidden := &cobra.Command{Use: "secret", Hidden: true, Aliases: []string{"s"}}
	tools.AddCommand(leafA)
	tools.AddCommand(leafB)
	tools.AddCommand(hidden)
	root.AddCommand(tools)

	merged := appendPluginInvocationEnv(
		nil,
		tools,
		sqlite.Plugin{CommandPath: "tools"},
		[]string{"/bin/mb", "tools", "--install"},
		"/cfg",
		[]string{"install"},
	)
	m := mapEnv(merged)

	var vis, hid []string
	if err := json.Unmarshal([]byte(m["MB_CTX_CHILD_COMMANDS"]), &vis); err != nil {
		t.Fatalf("CHILD: %v", err)
	}
	if err := json.Unmarshal([]byte(m["MB_CTX_HIDDEN_CHILD_COMMANDS"]), &hid); err != nil {
		t.Fatalf("HIDDEN: %v", err)
	}
	if got, want := vis, []string{"a", "b"}; !slicesEqual(got, want) {
		t.Errorf("MB_CTX_CHILD_COMMANDS = %v, want %v", got, want)
	}
	if got, want := hid, []string{"secret"}; !slicesEqual(got, want) {
		t.Errorf("MB_CTX_HIDDEN_CHILD_COMMANDS = %v, want %v", got, want)
	}

	var aliasEntries []childAliasEntry
	if err := json.Unmarshal([]byte(m["MB_CTX_CHILD_COMMAND_ALIASES"]), &aliasEntries); err != nil {
		t.Fatalf("ALIASES: %v", err)
	}
	if len(aliasEntries) != 2 {
		t.Fatalf("aliases len = %d, want 2: %#v", len(aliasEntries), aliasEntries)
	}
	// sorted by name: b before secret
	if aliasEntries[0].Name != "b" ||
		!slicesEqual(aliasEntries[0].Aliases, []string{"bee", "beta"}) {
		t.Errorf("first alias entry = %#v", aliasEntries[0])
	}
	if aliasEntries[1].Name != "secret" || !slicesEqual(aliasEntries[1].Aliases, []string{"s"}) {
		t.Errorf("second alias entry = %#v", aliasEntries[1])
	}

	var peers []string
	if err := json.Unmarshal([]byte(m["MB_CTX_PEER_COMMANDS"]), &peers); err != nil {
		t.Fatalf("peers: %v", err)
	}
	// tools is under mb; peers are other root commands — none registered here besides tools? Actually only tools under root, so peers empty
	if len(peers) != 0 {
		t.Errorf("MB_CTX_PEER_COMMANDS = %v, want [] (only tools under root in this tree)", peers)
	}
}

func mapEnv(merged []string) map[string]string {
	m := make(map[string]string, len(merged))
	for _, e := range merged {
		i := strings.IndexByte(e, '=')
		if i <= 0 {
			continue
		}
		m[e[:i]] = e[i+1:]
	}
	return m
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
