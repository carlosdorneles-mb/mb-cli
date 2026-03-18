package envgroup

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		group   string
		wantErr bool
	}{
		{"empty", "", true},
		{"dots", "..", true},
		{"slash", "a/b", true},
		{"spaces", "a b", true},
		{"ok staging", "staging", false},
		{"ok with dot", "v1.staging", false},
		{"ok underscore", "my_group", false},
		{"ok default", "default", false},
		{"too long", strings.Repeat("a", 65), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) err=%v wantErr=%v", tt.group, err, tt.wantErr)
			}
		})
	}
}

func TestFilePath(t *testing.T) {
	cfg := filepath.Join(string(filepath.Separator), "home", "u", ".config", "mb")
	p, err := FilePath(cfg, "staging")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(cfg, ".env.staging")
	if p != want {
		t.Errorf("got %q want %q", p, want)
	}
	_, err = FilePath(cfg, "../x")
	if err == nil {
		t.Error("expected error for invalid group")
	}
}
