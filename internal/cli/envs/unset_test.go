package envs

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEnvUnsetRemovesFromDefaultFile(t *testing.T) {
	d := testDeps(t)
	rootSet := NewCmd(d)
	rootSet.SetOut(&bytes.Buffer{})
	rootSet.SetErr(os.NewFile(0, os.DevNull))
	rootSet.SetArgs([]string{"set", "MYKEY=myval"})
	if err := rootSet.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}

	rootUnset := NewCmd(d)
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
