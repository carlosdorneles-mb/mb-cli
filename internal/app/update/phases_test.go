package update

import "testing"

func TestResolveUpdatePhases(t *testing.T) {
	tests := []struct {
		name                         string
		op, oc, os, ot               bool
		wantP, wantC, wantSys, wantT bool
	}{
		{"none", false, false, false, false, true, true, true, true},
		{"only plugins", true, false, false, false, true, false, false, false},
		{"only tools", false, false, false, true, false, false, false, true},
		{"only cli", false, true, false, false, false, true, false, false},
		{"only system", false, false, true, false, false, false, true, false},
		{"plugins+cli", true, true, false, false, true, true, false, false},
		{"all four", true, true, true, true, true, true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, c, s, tools := ResolveUpdatePhases(tt.op, tt.oc, tt.os, tt.ot)
			if p != tt.wantP || c != tt.wantC || s != tt.wantSys || tools != tt.wantT {
				t.Fatalf(
					"got plugins=%v cli=%v system=%v tools=%v want plugins=%v cli=%v system=%v tools=%v",
					p,
					c,
					s,
					tools,
					tt.wantP,
					tt.wantC,
					tt.wantSys,
					tt.wantT,
				)
			}
		})
	}
}
