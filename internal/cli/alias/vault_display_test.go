package alias

import "testing"

func TestNormalizeMbcliAliasVaultFlag(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"project", "", false},
		{"default", "", false},
		{"staging", "staging", false},
		{"project/staging", "staging", false},
		{"project/", "", true},
		{"project/staging/extra", "", true},
	}
	for _, tc := range tests {
		got, err := normalizeMbcliAliasVaultFlag(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%q: want error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestAliasMbcliLogicalVaultDisplay(t *testing.T) {
	if g := aliasMbcliLogicalVaultDisplay(""); g != "project" {
		t.Fatalf("empty: %q", g)
	}
	if g := aliasMbcliLogicalVaultDisplay("default"); g != "project" {
		t.Fatalf("default: %q", g)
	}
	if g := aliasMbcliLogicalVaultDisplay("staging"); g != "project/staging" {
		t.Fatalf("staging: %q", g)
	}
}

func TestAliasListVaultDisplay(t *testing.T) {
	if g := aliasListVaultDisplay("config", ""); g != "(nenhum)" {
		t.Fatalf("config empty: %q", g)
	}
	if g := aliasListVaultDisplay("config", "st"); g != "st" {
		t.Fatalf("config st: %q", g)
	}
	if g := aliasListVaultDisplay("project", ""); g != "project" {
		t.Fatalf("project root: %q", g)
	}
	if g := aliasListVaultDisplay("project", "staging"); g != "project/staging" {
		t.Fatalf("project nested: %q", g)
	}
}
