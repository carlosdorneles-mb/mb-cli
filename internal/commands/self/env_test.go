package self

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvListEmpty(t *testing.T) {
	d := testSelfDeps(t)
	root := newSelfEnvCmd(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
}

func TestEnvSetUnsetDefault(t *testing.T) {
	d := testSelfDeps(t)
	root := newSelfEnvCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"set", "MYKEY", "myval"})
	if err := root.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}

	b, err := os.ReadFile(d.Runtime.DefaultEnvPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "MYKEY") || !strings.Contains(string(b), "myval") {
		t.Errorf("env file: %s", b)
	}

	root2 := newSelfEnvCmd(d)
	root2.SetOut(&bytes.Buffer{})
	root2.SetErr(os.NewFile(0, os.DevNull))
	root2.SetArgs([]string{"unset", "MYKEY"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("unset: %v", err)
	}
	b, _ = os.ReadFile(d.Runtime.DefaultEnvPath)
	if strings.Contains(string(b), "MYKEY=myval") {
		t.Errorf("key should be removed: %s", b)
	}
}

func TestEnvSetGroup(t *testing.T) {
	d := testSelfDeps(t)
	root := newSelfEnvCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"set", "--group", "staging", "API", "https://x"})
	if err := root.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}
	groupPath := filepath.Join(d.Runtime.ConfigDir, ".env.staging")
	b, err := os.ReadFile(groupPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "API") {
		t.Errorf("group file: %s", b)
	}
}

func TestEnvListInvalidGroup(t *testing.T) {
	d := testSelfDeps(t)
	root := newSelfEnvCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"list", "--group", "grupo inválido"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEnvSetRequiresTwoArgs(t *testing.T) {
	d := testSelfDeps(t)
	root := newSelfEnvCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "only"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected arg error")
	}
}
