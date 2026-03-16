package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func Table(ctx context.Context, headers []string, rows [][]string, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		// Fallback plain output when gum is unavailable.
		fmt.Fprintln(out, strings.Join(headers, "\t"))
		for _, row := range rows {
			fmt.Fprintln(out, strings.Join(row, "\t"))
		}
		return nil
	}

	args := []string{"table", "--print"}
	args = append(args, "--separator", "\t", "--columns", strings.Join(headers, ","))

	var input bytes.Buffer
	for _, row := range rows {
		input.WriteString(strings.Join(row, "\t"))
		input.WriteString("\n")
	}

	cmd := exec.CommandContext(ctx, gumPath, args...)
	cmd.Stdin = &input
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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
