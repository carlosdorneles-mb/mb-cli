package envs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvSetPersistsToDefaultFile(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
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
}

func TestEnvSetGroup(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
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

func TestEnvSetSecretAndSecretOPMutuallyExclusive(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "K", "v", "--secret", "--secret-op"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for --secret together with --secret-op")
	}
}

func TestEnvSetRequiresTwoArgs(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "only"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected arg error")
	}
}

func TestEnvSetLogsToStderr(t *testing.T) {
	d := testDeps(t)
	var errBuf bytes.Buffer
	root := NewCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errBuf)
	root.SetArgs([]string{"set", "K", "v"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "K") {
		t.Errorf("expected log on stderr: %q", errBuf.String())
	}
}
