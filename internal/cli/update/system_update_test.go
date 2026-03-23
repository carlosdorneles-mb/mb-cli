package update

import (
	"context"
	"runtime"
	"slices"
	"strings"
	"testing"

	"mb/internal/shared/system"
)

func TestLinuxPackageEnv(t *testing.T) {
	t.Parallel()
	got := linuxPackageEnv("DEBIAN_FRONTEND=noninteractive")
	if !slices.Contains(got, "CI=1") {
		t.Fatal("expected CI=1 in env")
	}
	if !slices.Contains(got, "DEBIAN_FRONTEND=noninteractive") {
		t.Fatal("expected DEBIAN_FRONTEND=noninteractive in env")
	}
}

func TestRunSystemUpdateUnsupportedOSReturnsNil(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux":
		t.Skip("would invoke real package managers; only assert unsupported OS path here")
	}
	ctx := context.Background()
	var buf strings.Builder
	log := system.NewLogger(false, true, &buf)
	if err := RunSystemUpdate(ctx, log); err != nil {
		t.Fatalf("RunSystemUpdate: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "não suportado") && !strings.Contains(out, "SO não suportado") {
		t.Errorf("expected unsupported OS warning, got: %q", out)
	}
}
