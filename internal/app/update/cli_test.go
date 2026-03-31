package update

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
	"mb/internal/shared/system"
	"mb/internal/shared/version"
)

func testDepsForUpdateCLI(t *testing.T) deps.Dependencies {
	t.Helper()
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "cache.db")
	pluginsDir := filepath.Join(tmp, "plugins")
	configDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("mkdir plugins: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			PluginsDir: pluginsDir,
			ConfigDir:  configDir,
		},
	}
	return deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
	)
}

func TestSelfUpdateConfigFromDeps(t *testing.T) {
	t.Parallel()
	d := deps.Dependencies{AppConfig: config.AppConfig{UpdateRepo: "  my/repo  "}}
	cfg := SelfUpdateConfigFromDeps(d)
	if cfg == nil {
		t.Fatal("nil config")
	}
	if cfg.Repo != "my/repo" {
		t.Errorf("Repo = %q, want my/repo", cfg.Repo)
	}

	empty := SelfUpdateConfigFromDeps(deps.Dependencies{AppConfig: config.AppConfig{}})
	if empty.Repo != "" {
		t.Errorf("empty UpdateRepo: Repo = %q, want \"\"", empty.Repo)
	}
}

func TestLogInfoLinesSkipsBlankLines(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var buf strings.Builder
	log := system.NewLogger(false, false, &buf)
	LogInfoLines(ctx, log, "  a  \n\n  b  \n")
	out := buf.String()
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected both lines logged, got: %q", out)
	}
}

func TestRunCLIUpdateNonReleaseBuildReturnsNil(t *testing.T) {
	if version.IsReleaseBuild() {
		t.Skip("test binary has embedded Version; skip release path that calls network")
	}
	ctx := context.Background()
	d := testDepsForUpdateCLI(t)
	var buf strings.Builder
	log := system.NewLogger(false, false, &buf)
	if err := RunCLIUpdate(ctx, d, log); err != nil {
		t.Fatalf("RunCLIUpdate: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Atualizando MB CLI") {
		t.Errorf("expected intro log, got: %q", out)
	}
	if !strings.Contains(out, "release oficial") {
		t.Errorf("expected non-release hint in output: %q", out)
	}
}
