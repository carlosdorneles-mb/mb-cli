package envs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvUnsetMbcliYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	y := filepath.Join(tmp, "mbcli.yaml")
	if err := os.WriteFile(y, []byte("envs:\n  RM: \"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MBCLI_YAML_PATH", y)
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"unset", "RM", "--mbcli-yaml", "--yes"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unset: %v", err)
	}
	b, err := os.ReadFile(y)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "RM:") {
		t.Fatalf("key should be removed: %s", string(b))
	}
}

func TestEnvUnsetRemovesFromDefaultFile(t *testing.T) {
	d := testDeps(t)
	rootSet := NewCmd(testListServiceForDeps(t, d), d)
	rootSet.SetOut(&bytes.Buffer{})
	rootSet.SetErr(os.NewFile(0, os.DevNull))
	rootSet.SetArgs([]string{"set", "MYKEY=myval"})
	if err := rootSet.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}

	rootUnset := NewCmd(testListServiceForDeps(t, d), d)
	rootUnset.SetOut(&bytes.Buffer{})
	rootUnset.SetErr(os.NewFile(0, os.DevNull))
	rootUnset.SetArgs([]string{"unset", "MYKEY"})
	if err := rootUnset.Execute(); err != nil {
		t.Fatalf("unset: %v", err)
	}
	b, _ := os.ReadFile(d.Runtime.DefaultEnvPath)
	if strings.Contains(string(b), "MYKEY=myval") {
		t.Errorf("key should be removed: %s", b)
	}
}
