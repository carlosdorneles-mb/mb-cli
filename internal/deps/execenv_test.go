package deps

import (
	"strings"
	"testing"
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
		rt        *RuntimeConfig
		wantVerb  bool
		wantQuiet bool
	}{
		{"both false", &RuntimeConfig{Verbose: false, Quiet: false}, false, false},
		{"verbose only", &RuntimeConfig{Verbose: true, Quiet: false}, true, false},
		{"quiet only", &RuntimeConfig{Verbose: false, Quiet: true}, false, true},
		{"both true", &RuntimeConfig{Verbose: true, Quiet: true}, true, true},
		{"nil rt", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := AppendVerbosityEnv([]string{"FOO=bar"}, tt.rt)
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
