package system

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"mb/internal/shared/ui"
)

// Confirm asks yes/no via gum confirm when gum is on PATH; otherwise prompt on out and read a line from in (y/yes = true).
func Confirm(ctx context.Context, prompt string, in io.Reader, out io.Writer) (bool, error) {
	if out == nil {
		out = os.Stderr
	}
	if in == nil {
		in = os.Stdin
	}
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		return confirmFallback(prompt, in, out)
	}
	cmd := exec.CommandContext(ctx, gumPath,
		"confirm", prompt,
		"--affirmative", "Sim",
		"--negative", "Não",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = ui.PrependGumThemeDefaults(os.Environ())
	err = cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

func confirmFallback(prompt string, in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprintf(out, "%s (y/N): ", prompt)
	var answer string
	_, err := fmt.Fscanln(in, &answer)
	if err != nil && err.Error() != "unexpected newline" {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}
