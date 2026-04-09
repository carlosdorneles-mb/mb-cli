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
}

func TestPeerCommandsJSON_nilCmd(t *testing.T) {
	t.Parallel()
	if got := peerCommandsJSON(nil); got != "[]" {
		t.Errorf("peerCommandsJSON(nil) = %q", got)
	}
}
