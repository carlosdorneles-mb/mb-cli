package envs

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestSortedKeys(t *testing.T) {
	t.Parallel()
	if got := sortedKeys(nil); len(got) != 0 {
		t.Errorf("sortedKeys(nil) = %v, want empty", got)
	}
	if got := sortedKeys(map[string]string{}); len(got) != 0 {
		t.Errorf("sortedKeys({}) = %v, want empty", got)
	}
	got := sortedKeys(map[string]string{"z": "1", "a": "2", "m": "3"})
	want := []string{"a", "m", "z"}
	if !slices.Equal(got, want) {
		t.Errorf("sortedKeys = %v, want %v", got, want)
	}
}

func TestEnvTargetPathDefault(t *testing.T) {
	t.Parallel()
	d := testDeps(t)
	p, err := envTargetPath(d, "")
	if err != nil {
		t.Fatal(err)
	}
	if p != d.Runtime.DefaultEnvPath {
		t.Errorf("path = %q, want %q", p, d.Runtime.DefaultEnvPath)
	}
}

func TestEnvTargetPathGroup(t *testing.T) {
	t.Parallel()
	d := testDeps(t)
	p, err := envTargetPath(d, "staging")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(d.Runtime.ConfigDir, ".env.staging")
	if p != want {
		t.Errorf("path = %q, want %q", p, want)
	}
}

func TestEnvTargetPathInvalidGroup(t *testing.T) {
	t.Parallel()
	d := testDeps(t)
	_, err := envTargetPath(d, "grupo inválido")
	if err == nil {
		t.Fatal("expected error for invalid group name")
	}
}
