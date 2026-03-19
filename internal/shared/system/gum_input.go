package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Input(ctx context.Context, prompt string) (string, error) {
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		return "", fmt.Errorf("gum not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, gumPath, "input", "--prompt", prompt+" ")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
