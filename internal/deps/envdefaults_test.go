package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildEnvFileValues_DefaultOnly(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p, []byte("A=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: p,
		},
	}
	m, err := BuildEnvFileValues(rt)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "1" {
		t.Errorf("A=%q", m["A"])
	}
}

func TestBuildEnvFileValues_GroupOverlay(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("A=1\nX=0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	grp := filepath.Join(tmp, ".env.staging")
	if err := os.WriteFile(grp, []byte("A=3\nB=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
		EnvGroup: "staging",
	}
	m, err := BuildEnvFileValues(rt)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "3" || m["B"] != "2" || m["X"] != "0" {
		t.Fatalf("merged=%v", m)
	}
}

func TestBuildEnvFileValues_InvalidEnvGroup(t *testing.T) {
	tmp := t.TempDir()
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: filepath.Join(tmp, "env.defaults"),
		},
		EnvGroup: "../x",
	}
	_, err := BuildEnvFileValues(rt)
	if err == nil {
		t.Fatal("expected error")
	}
}
