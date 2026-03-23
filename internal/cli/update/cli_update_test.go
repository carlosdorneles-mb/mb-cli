package update

import (
	"context"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/shared/config"
	"mb/internal/shared/system"
	"mb/internal/shared/version"
)

func TestSelfupdateFromAppConfig(t *testing.T) {
	t.Parallel()
	d := deps.Dependencies{AppConfig: config.AppConfig{UpdateRepo: "  my/repo  "}}
	cfg := selfupdateFromAppConfig(d)
	if cfg == nil {
		t.Fatal("nil config")
	}
	if cfg.Repo != "my/repo" {
		t.Errorf("Repo = %q, want my/repo", cfg.Repo)
	}

	empty := selfupdateFromAppConfig(deps.Dependencies{AppConfig: config.AppConfig{}})
	if empty.Repo != "" {
		t.Errorf("empty UpdateRepo: Repo = %q, want \"\"", empty.Repo)
	}
}

func TestLogInfoLinesSkipsBlankLines(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var buf strings.Builder
	log := system.NewLogger(false, false, &buf)
	logInfoLines(ctx, log, "  a  \n\n  b  \n")
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
	d := testUpdateDeps(t)
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
