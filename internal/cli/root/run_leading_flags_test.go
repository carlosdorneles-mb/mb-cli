package root

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

func testRootDepsRunFlags(t *testing.T, cfgDir string, rt *deps.RuntimeConfig) deps.Dependencies {
	t.Helper()
	cachePath := filepath.Join(cfgDir, "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	pluginsDir := filepath.Join(cfgDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir plugins: %v", err)
	}
	rt.PluginsDir = pluginsDir
	return deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)
}

func TestRootRun_LeadingEnvInjectsMergedEnv(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, "mb")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{ConfigDir: cfgDir, DefaultEnvPath: def},
	}
	d := testRootDepsRunFlags(t, cfgDir, rt)
	fsys, git, shell, layout := testRootInfraPorts(t)
	root := NewRootCmd(d, fsys, git, shell, layout)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origOut := os.Stdout
	os.Stdout = w
	root.SetArgs([]string{"run", "-e", "MB_RUN_LEAD=ok", "sh", "-c", `printf '%s' "$MB_RUN_LEAD"`})
	execErr := root.Execute()
	_ = w.Close()
	os.Stdout = origOut
	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	body, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "ok" {
		t.Fatalf("stdout=%q want ok", got)
	}
}

func TestRootRun_LeadingEnvVault(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, "mb")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(cfgDir, ".env.gprod"),
		[]byte("MB_RUN_GROUP=from_overlay\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{ConfigDir: cfgDir, DefaultEnvPath: def},
	}
	d := testRootDepsRunFlags(t, cfgDir, rt)
	fsys, git, shell, layout := testRootInfraPorts(t)
	root := NewRootCmd(d, fsys, git, shell, layout)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origOut := os.Stdout
	os.Stdout = w
	root.SetArgs([]string{"run", "--env-vault", "gprod", "sh", "-c", `printf '%s' "$MB_RUN_GROUP"`})
	execErr := root.Execute()
	_ = w.Close()
	os.Stdout = origOut
	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	body, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "from_overlay" {
		t.Fatalf("stdout=%q want from_overlay", got)
	}
}

func TestRootRun_CombineRootAndRunEnvFlags(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, "mb")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{ConfigDir: cfgDir, DefaultEnvPath: def},
	}
	d := testRootDepsRunFlags(t, cfgDir, rt)
	fsys, git, shell, layout := testRootInfraPorts(t)
	root := NewRootCmd(d, fsys, git, shell, layout)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origOut := os.Stdout
	os.Stdout = w
	root.SetArgs([]string{"-e", "A=1", "run", "-e", "B=2", "sh", "-c", `printf '%s' "$A$B"`})
	execErr := root.Execute()
	_ = w.Close()
	os.Stdout = origOut
	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	body, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "12" {
		t.Fatalf("stdout=%q want 12", got)
	}
}

func TestRootRun_GrepStyleFlagsAfterCommand(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, "mb")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{ConfigDir: cfgDir, DefaultEnvPath: def},
	}
	d := testRootDepsRunFlags(t, cfgDir, rt)
	fsys, git, shell, layout := testRootInfraPorts(t)
	root := NewRootCmd(d, fsys, git, shell, layout)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origOut := os.Stdout
	os.Stdout = w
	root.SetArgs([]string{"run", "grep", "-c", "^a", filepath.Join(tmp, "f")})
	if err := os.WriteFile(filepath.Join(tmp, "f"), []byte("a\nb\na\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	execErr := root.Execute()
	_ = w.Close()
	os.Stdout = origOut
	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	body, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(body)); got != "2" {
		t.Fatalf("grep -c output=%q want 2", got)
	}
}
