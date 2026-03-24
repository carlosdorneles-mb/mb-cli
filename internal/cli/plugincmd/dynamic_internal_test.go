package plugincmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

func TestAppendVerbosityEnv(t *testing.T) {
	contains := func(env []string, key string) bool {
		for _, e := range env {
			if e == key+"=1" {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name      string
		rt        *deps.RuntimeConfig
		wantVerb  bool
		wantQuiet bool
	}{
		{"both false", &deps.RuntimeConfig{Verbose: false, Quiet: false}, false, false},
		{"verbose only", &deps.RuntimeConfig{Verbose: true, Quiet: false}, true, false},
		{"quiet only", &deps.RuntimeConfig{Verbose: false, Quiet: true}, false, true},
		{"both true", &deps.RuntimeConfig{Verbose: true, Quiet: true}, true, true},
		{"nil rt", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := deps.AppendVerbosityEnv([]string{"FOO=bar"}, tt.rt)
			if got := contains(merged, "MB_VERBOSE"); got != tt.wantVerb {
				t.Errorf(
					"AppendVerbosityEnv() MB_VERBOSE present = %v, want %v (merged: %s)",
					got,
					tt.wantVerb,
					strings.Join(merged, " "),
				)
			}
			if got := contains(merged, "MB_QUIET"); got != tt.wantQuiet {
				t.Errorf(
					"AppendVerbosityEnv() MB_QUIET present = %v, want %v (merged: %s)",
					got,
					tt.wantQuiet,
					strings.Join(merged, " "),
				)
			}
			if !tt.wantVerb && !tt.wantQuiet && tt.rt != nil {
				if len(merged) != 1 || merged[0] != "FOO=bar" {
					t.Errorf(
						"AppendVerbosityEnv() should not add entries when both false, got %v",
						merged,
					)
				}
			}
		})
	}
}

func TestParseRootVerbosityFlags(t *testing.T) {
	var verbose, quiet bool
	root := &cobra.Command{Use: "mb"}
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "")
	child := &cobra.Command{Use: "hello"}
	root.AddCommand(child)

	tests := []struct {
		name          string
		args          []string
		wantVerbose   bool
		wantQuiet     bool
		wantRemaining []string
	}{
		{"-v consumes and sets verbose", []string{"-v"}, true, false, []string{}},
		{"-q consumes and sets quiet", []string{"-q"}, false, true, []string{}},
		{"-v -q both set", []string{"-v", "-q"}, true, true, []string{}},
		{"no flags", []string{}, false, false, []string{}},
		{"-v then positional", []string{"-v", "foo"}, true, false, []string{"foo"}},
		{"positional then -v", []string{"foo", "-v"}, true, false, []string{"foo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verbose, quiet = false, false
			remaining := parseRootVerbosityFlags(child, tt.args)
			if verbose != tt.wantVerbose {
				t.Errorf("verbose = %v, want %v", verbose, tt.wantVerbose)
			}
			if quiet != tt.wantQuiet {
				t.Errorf("quiet = %v, want %v", quiet, tt.wantQuiet)
			}
			if !reflect.DeepEqual(remaining, tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}

func TestNewLeafCommand_FlagsWithReadmeNoPanicWhenPluginUsesR(t *testing.T) {
	flagsJSON := `{"run":{"type":"long","short":"r","entrypoint":"run.sh","description":"run"}}`
	plugin := sqlite.Plugin{
		CommandPath: "dev/bump",
		CommandName: "bump",
		FlagsJSON:   flagsJSON,
		ReadmePath:  "/tmp/mb-readme-test.md",
	}
	d := deps.NewDependencies(&deps.RuntimeConfig{}, config.AppConfig{}, nil, nil, nil)
	cmd := newLeafCommand("bump", plugin, d, "/tmp", false, nil, nil)
	rf := cmd.Flags().Lookup("readme")
	if rf == nil {
		t.Fatal("readme flag missing")
	}
	if rf.Shorthand != "r" {
		t.Fatalf("readme should keep -r (MB registado antes dos flags do plugin), got shorthand %q", rf.Shorthand)
	}
	runF := cmd.Flags().Lookup("run")
	if runF == nil || runF.Shorthand != "" {
		t.Fatalf("run flag should be long-only when -r fica com --readme, got %#v", runF)
	}
}

func TestNewLeafCommand_ReservedRootShorthandDropped(t *testing.T) {
	flagsJSON := `{"watch":{"type":"long","short":"v","entrypoint":"w.sh","description":"w"}}`
	plugin := sqlite.Plugin{
		CommandPath: "p/x",
		CommandName: "x",
		FlagsJSON:   flagsJSON,
	}
	d := deps.NewDependencies(&deps.RuntimeConfig{}, config.AppConfig{}, nil, nil, nil)
	global := map[string]struct{}{"v": {}}
	cmd := newLeafCommand("x", plugin, d, "/tmp", false, nil, global)
	f := cmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("watch flag missing")
	}
	if f.Shorthand != "" {
		t.Fatalf("want empty shorthand when manifest requests reserved v, got %q", f.Shorthand)
	}
}
