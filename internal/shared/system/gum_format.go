package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// GumStyleBox wraps text in a styled box with rounded borders using gum style.
// If gum is not found, prints the content as-is.
func GumStyleBox(ctx context.Context, content string, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}

	gumPath, err := exec.LookPath("gum")
	if err != nil {
		// Fallback: print as-is
		fmt.Fprint(out, content)
		return nil
	}

	args := []string{
		"style",
		"--border", "rounded",
		"--padding", "1 3",
		"--margin", "1 0",
		"--border-foreground", "212",
	}

	cmd := exec.CommandContext(ctx, gumPath, args...)
	cmd.Stdin = bytes.NewBufferString(content)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

// GumStyleBoxWithTitle creates a styled box with a title header.
func GumStyleBoxWithTitle(ctx context.Context, title string, content string, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}

	fullContent := fmt.Sprintf("%s\n\n%s", title, content)
	return GumStyleBox(ctx, fullContent, out)
}
