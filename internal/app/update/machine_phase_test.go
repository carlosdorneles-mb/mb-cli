package update

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/shared/system"
)

func TestFindMachineUpdateCmd(t *testing.T) {
	root := &cobra.Command{Use: "mb"}
	if findMachineUpdateCmd(root) {
		t.Fatal("expected false with no commands")
	}
	machine := &cobra.Command{Use: "machine"}
	machine.AddCommand(
		&cobra.Command{Use: "update", RunE: func(*cobra.Command, []string) error { return nil }},
	)
	root.AddCommand(machine)
	if !findMachineUpdateCmd(root) {
		t.Fatal("expected true with machine update")
	}
}

func TestRunMachineUpdatePhaseSkipWarnsWhenExclusive(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{Use: "mb"}
	updateCmd := &cobra.Command{Use: "update"}
	root.AddCommand(updateCmd)
	var buf strings.Builder
	log := system.NewLogger(false, true, &buf)
	if err := RunMachineUpdatePhase(ctx, updateCmd, log, true); err != nil {
		t.Fatalf("RunMachineUpdatePhase: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "machine update indisponível") {
		t.Errorf("expected warn about machine update, got: %q", out)
	}
}

func TestRunMachineUpdatePhaseSkipSilentWhenNotExclusive(t *testing.T) {
	ctx := context.Background()
	root := &cobra.Command{Use: "mb"}
	updateCmd := &cobra.Command{Use: "update"}
	root.AddCommand(updateCmd)
	var buf strings.Builder
	log := system.NewLogger(false, true, &buf)
	if err := RunMachineUpdatePhase(ctx, updateCmd, log, false); err != nil {
		t.Fatalf("RunMachineUpdatePhase: %v", err)
	}
	if strings.Contains(buf.String(), "machine update indisponível") {
		t.Errorf("did not expect warn when not exclusive, got: %q", buf.String())
	}
}
