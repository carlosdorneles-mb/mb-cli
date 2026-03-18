package env

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
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

func TestEnvListJSON(t *testing.T) {
	d := testDeps(t)
	if err := os.WriteFile(
		d.Runtime.DefaultEnvPath,
		[]byte("FOO=bar\nBAZ=qux\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	root := NewCmd(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"list", "-J"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &got); err != nil {
		t.Fatalf("json: %v out=%q", err, out.String())
	}
	if len(got) != 2 || got["FOO"] != "bar" || got["BAZ"] != "qux" {
		t.Fatalf("got %v", got)
	}
}

func TestEnvListText(t *testing.T) {
	d := testDeps(t)
	if err := os.WriteFile(d.Runtime.DefaultEnvPath, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := NewCmd(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"list", "--text"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if strings.TrimSpace(out.String()) != "FOO=bar" {
		t.Fatalf("out=%q", out.String())
	}
}

func TestEnvListJSONAndTextMutuallyExclusive(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"list", "-J", "-T"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for -J and -T together")
	}
}
