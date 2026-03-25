package plugincmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

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
