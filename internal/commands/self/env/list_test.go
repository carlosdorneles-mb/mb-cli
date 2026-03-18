package env

import (
	"bytes"
	"os"
	"testing"
)

func TestEnvListEmpty(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
}

func TestEnvListInvalidGroup(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"list", "--group", "grupo inválido"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
}
