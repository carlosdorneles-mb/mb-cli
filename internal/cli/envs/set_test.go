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
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"set", "MYKEY=myval"})
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

func TestEnvSetVaultProjectReserved(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "--vault", "project", "X=1"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for reserved vault name project")
	}
}

func TestEnvSetVault(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"set", "--vault", "staging", "API=https://x"})
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
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "K", "v", "--secret", "--secret-op"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for --secret together with --secret-op")
	}
}

func TestEnvSetRequiresKeyEqualsValue(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "onlykey"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected arg error")
	}
}

func TestEnvSetLogsToStderr(t *testing.T) {
	d := testDeps(t)
	var errBuf bytes.Buffer
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errBuf)
	// Sem gum no PATH o logger usa fmt.Fprintf e a mensagem cai no buffer (gum log pode falhar sem TTY).
	t.Setenv("PATH", t.TempDir())
	root.SetArgs([]string{"set", "K=v"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "K") {
		t.Errorf("expected log on stderr: %q", errBuf.String())
	}
}

func TestParseEnvSetArgsBareKeyRequiresSecretMode(t *testing.T) {
	_, err := parseEnvSetArgs([]string{"ONLYKEY"}, false)
	if err == nil {
		t.Fatal("expected error without secret mode")
	}
	pairs, err := parseEnvSetArgs([]string{"ONLYKEY"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 || pairs[0].key != "ONLYKEY" || pairs[0].needsPrompt != true ||
		pairs[0].value != "" {
		t.Fatalf("pairs=%+v", pairs)
	}
}

func TestParseEnvSetArgsMixedSecret(t *testing.T) {
	pairs, err := parseEnvSetArgs([]string{"A", "B=2", "C"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 3 {
		t.Fatalf("len=%d pairs=%+v", len(pairs), pairs)
	}
	if pairs[0].key != "A" || !pairs[0].needsPrompt {
		t.Fatalf("pair0=%+v", pairs[0])
	}
	if pairs[1].key != "B" || pairs[1].value != "2" || pairs[1].needsPrompt {
		t.Fatalf("pair1=%+v", pairs[1])
	}
	if pairs[2].key != "C" || !pairs[2].needsPrompt {
		t.Fatalf("pair2=%+v", pairs[2])
	}
}

func TestParseEnvSetArgsEmptyValueWithEquals(t *testing.T) {
	pairs, err := parseEnvSetArgs([]string{"EMPTY="}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 || pairs[0].key != "EMPTY" || pairs[0].value != "" || pairs[0].needsPrompt {
		t.Fatalf("pairs=%+v", pairs)
	}
}

func TestEnvSetMbcliYAMLWritesYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	y := filepath.Join(tmp, "mbcli.yaml")
	t.Setenv("MBCLI_YAML_PATH", y)
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(os.NewFile(0, os.DevNull))
	root.SetArgs([]string{"set", "FROMCLI=ok", "--mbcli-yaml", "--yes"})
	if err := root.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}
	b, err := os.ReadFile(y)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "FROMCLI") || !strings.Contains(s, "ok") {
		t.Fatalf("mbcli.yaml: %s", s)
	}
}

func TestEnvSetMbcliYAMLMutuallyExclusiveWithSecret(t *testing.T) {
	d := testDeps(t)
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"set", "K=v", "--mbcli-yaml", "--secret"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected cobra mutual exclusive error")
	}
}

func TestEnvSetSecretInlineValueEmitsSecurityWarning(t *testing.T) {
	d := testDepsWithSecretStore(t, make(memorySecretStore))
	var errBuf bytes.Buffer
	root := NewCmd(testListServiceForDeps(t, d), d)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errBuf)
	t.Setenv("PATH", t.TempDir())
	root.SetArgs([]string{"set", "TOKEN=secretval", "--secret"})
	if err := root.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}
	out := errBuf.String()
	if !strings.Contains(out, "não é seguro") || !strings.Contains(out, "histórico") {
		t.Fatalf("expected security warning on stderr, got: %q", out)
	}
}
