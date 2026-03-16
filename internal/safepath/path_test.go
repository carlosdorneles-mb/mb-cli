package safepath

import (
	"path/filepath"
	"testing"
)

func TestPathUnderDir(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "plugin")
	script := filepath.Join(sub, "run.sh")

	tests := []struct {
		path string
		dir  string
		want bool
	}{
		{script, sub, true},
		{sub, sub, true},
		{sub, tmp, true},
		{filepath.Join(sub, "a", "b"), sub, true},
		{tmp, sub, false},
		{filepath.Join(sub, "..", "run.sh"), sub, false},
		{filepath.Join(sub, ".."), sub, false},
		{filepath.Join(sub, "..", "other", "x"), sub, false},
	}
	for _, tt := range tests {
		got, err := PathUnderDir(tt.path, tt.dir)
		if err != nil {
			t.Errorf("PathUnderDir(%q, %q): %v", tt.path, tt.dir, err)
			continue
		}
		if got != tt.want {
			t.Errorf("PathUnderDir(%q, %q) = %v, want %v", tt.path, tt.dir, got, tt.want)
		}
	}
}

func TestValidateUnderDir(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "plugin")
	script := filepath.Join(sub, "run.sh")

	if err := ValidateUnderDir(script, sub); err != nil {
		t.Errorf("ValidateUnderDir(script, sub): %v", err)
	}
	if err := ValidateUnderDir(filepath.Join(sub, "..", "x"), sub); err != ErrPathOutsideDir {
		t.Errorf("ValidateUnderDir(escape): got %v, want ErrPathOutsideDir", err)
	}
}
