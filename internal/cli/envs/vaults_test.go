package envs

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvVaultsTable(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"vaults"})
	if err := root.Execute(); err != nil {
		t.Fatalf("vaults: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected output")
	}
}

func TestEnvVaultsJSON(t *testing.T) {
	d := testDeps(t)
	if err := os.WriteFile(
		filepath.Join(d.Runtime.ConfigDir, ".env.staging"),
		[]byte("A=b\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	root := NewCmd(testListServiceForDeps(t, d), d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"vaults", "-J"})
	if err := root.Execute(); err != nil {
		t.Fatalf("vaults: %v", err)
	}
	var got []struct {
		Vault string `json:"vault"`
		Path  string `json:"path"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &got); err != nil {
		t.Fatalf("json: %v out=%q", err, out.String())
	}
	if len(got) != 2 {
		t.Fatalf("len=%d got=%+v", len(got), got)
	}
	if got[0].Vault != "default" || got[1].Vault != "staging" {
		t.Fatalf("order/vaults: %+v", got)
	}
}
