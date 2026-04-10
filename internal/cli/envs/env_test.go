package envs

import (
	"maps"
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCmd(t *testing.T) {
	t.Parallel()
	d := testDeps(t)
	cmd := NewCmd(testListServiceForDeps(t, d), d)

	if cmd.Use != "envs" {
		t.Errorf("Use = %q, want envs", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected non-empty Short")
	}

	subs := cmd.Commands()
	names := make([]string, len(subs))
	for i, c := range subs {
		names[i] = c.Name()
	}
	slices.Sort(names)
	want := []string{"list", "set", "unset", "vaults"}
	if !slices.Equal(names, want) {
		t.Errorf("subcommands = %v, want %v", names, want)
	}

	list := findSub(subs, "list")
	if list == nil {
		t.Fatal("missing list")
	}
	wantAliases := map[string]bool{"ls": true, "l": true}
	gotAliases := make(map[string]bool)
	for _, a := range list.Aliases {
		gotAliases[a] = true
	}
	if !maps.Equal(gotAliases, wantAliases) {
		t.Errorf("list aliases = %v, want ls and l", list.Aliases)
	}
}

func findSub(subs []*cobra.Command, name string) *cobra.Command {
	for _, c := range subs {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
