package alias

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/config"
)

func testDepsForAlias(t *testing.T, cfgDir string) deps.Dependencies {
	t.Helper()
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      cfgDir,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(cfgDir, "plugins"),
		},
	}
	return deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
}

func TestAliasSet_updateWithoutYes_nonInteractiveErrors(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "old"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "--", "echo", "new"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("expected error mentioning --yes, got %v", err)
	}

	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	sk := alib.StoreKey("", "dev")
	if len(got.Aliases[sk].Command) != 2 || got.Aliases[sk].Command[1] != "old" {
		t.Fatalf("alias should be unchanged, got %#v", got.Aliases[sk])
	}
}

func TestAliasSet_sameCommandNewVault_withYes_updatesVault(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "hi"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "--yes", "--vault", "staging", "--", "echo", "hi"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	root := got.Aliases[alib.StoreKey("", "dev")]
	if root.EnvVault != "" || len(root.Command) != 2 || root.Command[1] != "hi" {
		t.Fatalf("root dev=%+v", root)
	}
	st := got.Aliases[alib.StoreKey("staging", "dev")]
	if st.EnvVault != "staging" || len(st.Command) != 2 || st.Command[1] != "hi" {
		t.Fatalf("staging dev=%+v", st)
	}
	if len(got.Aliases) != 2 {
		t.Fatalf("want two slots (nome+vault), got %d", len(got.Aliases))
	}
}

func TestAliasSet_idempotentNoPrompt_nonInteractiveOK(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "x"}, EnvVault: "st"},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "--vault", "st", "--", "echo", "x"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	e := got.Aliases[alib.StoreKey("st", "dev")]
	if e.EnvVault != "st" || len(e.Command) != 2 || e.Command[1] != "x" {
		t.Fatalf("got %#v", e)
	}
}

func TestAliasSet_updateWithYes_writes(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "old"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "--yes", "--", "echo", "new"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	sk := alib.StoreKey("", "dev")
	if len(got.Aliases[sk].Command) != 2 || got.Aliases[sk].Command[1] != "new" {
		t.Fatalf("expected updated command, got %#v", got.Aliases[sk])
	}
}

func TestAliasSet_createWithoutConfirm_nonInteractiveOK(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	emptyFile := &alib.File{Version: 1, Aliases: map[string]alib.Entry{}}
	if err := alib.Save(path, emptyFile); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"newalias", "--", "true"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	e, ok := got.Aliases[alib.StoreKey("", "newalias")]
	if !ok || len(e.Command) != 1 || e.Command[0] != "true" {
		t.Fatalf("expected new alias, got %#v", got.Aliases)
	}
}

func TestAliasUnset_withoutYes_nonInteractiveErrors(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "x"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newUnsetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("expected error mentioning --yes, got %v", err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Aliases[alib.StoreKey("", "dev")]; !ok {
		t.Fatal("alias should still exist")
	}
}

func TestAliasUnset_withYes_removes(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "x"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newUnsetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "--yes"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Aliases[alib.StoreKey("", "dev")]; ok {
		t.Fatal("expected alias removed")
	}
}

func TestAliasUnset_withYes_removesTwo(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"api":    {Command: []string{"echo", "api"}},
			"worker": {Command: []string{"echo", "worker"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newUnsetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"api", "worker", "--yes"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Aliases) != 0 {
		t.Fatalf("expected no aliases, got %#v", got.Aliases)
	}
}

func TestAliasUnset_missingOneErrorsAndLeavesFileUnchanged(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "x"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newUnsetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "ghost", "--yes"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	err := cmd.Execute()
	if err == nil ||
		!strings.Contains(err.Error(), "aliases inexistentes") ||
		!strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected error about missing alias, got %v", err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Aliases[alib.StoreKey("", "dev")]; !ok {
		t.Fatal("alias dev should still exist")
	}
}

func TestAliasUnset_dedupeWithYes_removesOnce(t *testing.T) {
	cfgDir := t.TempDir()
	path := alib.FilePath(cfgDir)
	initial := &alib.File{
		Version: 1,
		Aliases: map[string]alib.Entry{
			"dev": {Command: []string{"echo", "x"}},
		},
	}
	if err := alib.Save(path, initial); err != nil {
		t.Fatal(err)
	}

	d := testDepsForAlias(t, cfgDir)
	cmd := newUnsetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"dev", "dev", "--yes"})
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetOut(bytes.NewBuffer(nil))

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := alib.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Aliases[alib.StoreKey("", "dev")]; ok {
		t.Fatal("expected alias removed")
	}
}
