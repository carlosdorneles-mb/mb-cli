package self

import (
	"bytes"
	"strings"
	"testing"

	"mb/internal/commands/config"
	"mb/internal/version"
)

func TestSelfUpdateQuietNonReleaseNoStdout(t *testing.T) {
	orig := version.Version
	t.Cleanup(func() { version.Version = orig })
	version.Version = ""

	rt := &config.RuntimeConfig{Quiet: true}
	deps := config.Dependencies{Runtime: rt}

	var buf bytes.Buffer
	cmd := newSelfUpdateCmd(deps)
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("quiet: expected empty stdout, got %q", buf.String())
	}
}

func TestSelfUpdateNonQuietNonReleasePrints(t *testing.T) {
	orig := version.Version
	t.Cleanup(func() { version.Version = orig })
	version.Version = ""

	rt := &config.RuntimeConfig{Quiet: false}
	deps := config.Dependencies{Runtime: rt}

	var buf bytes.Buffer
	cmd := newSelfUpdateCmd(deps)
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "release oficial") {
		t.Fatalf("expected non-release message, got %q", buf.String())
	}
}
