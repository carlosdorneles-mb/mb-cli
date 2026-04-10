package envs

import (
	"path/filepath"
	"testing"
)

func TestTargetPath_rejectsReservedProject(t *testing.T) {
	tmp := t.TempDir()
	paths := Paths{
		DefaultEnvPath: filepath.Join(tmp, "env.defaults"),
		ConfigDir:      tmp,
	}
	_, err := TargetPath(paths, "project")
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = TargetPath(paths, "project/staging")
	if err == nil {
		t.Fatal("expected error for project/ prefix")
	}
}
