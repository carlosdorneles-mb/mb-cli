package self

import (
	"bytes"
	"strings"
	"testing"

	"mb/internal/deps"
	"mb/internal/version"
)

func TestSelfUpdateQuietNonReleaseNoOutput(t *testing.T) {
	orig := version.Version
	t.Cleanup(func() { version.Version = orig })
	version.Version = ""

	rt := &deps.RuntimeConfig{Quiet: true}
	d := deps.Dependencies{Runtime: rt}

	var stdout, stderr bytes.Buffer
	cmd := newSelfUpdateCmd(d)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	t.Setenv("PATH", t.TempDir())
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("quiet: expected empty stdout, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("quiet: expected empty stderr, got %q", stderr.String())
	}
}

func TestSelfUpdateNonQuietNonReleasePrints(t *testing.T) {
	orig := version.Version
	t.Cleanup(func() { version.Version = orig })
	version.Version = ""

	rt := &deps.RuntimeConfig{Quiet: false}
	d := deps.Dependencies{Runtime: rt}

	var stdout, stderr bytes.Buffer
	cmd := newSelfUpdateCmd(d)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	t.Setenv("PATH", t.TempDir())
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	combined := stderr.String() + stdout.String()
	if !strings.Contains(combined, "release oficial") {
		t.Fatalf(
			"expected non-release message, stderr=%q stdout=%q",
			stderr.String(),
			stdout.String(),
		)
	}
}
