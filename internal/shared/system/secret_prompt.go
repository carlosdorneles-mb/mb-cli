package system

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"

	"mb/internal/shared/ui"
)

// PromptSecretValue reads a secret value for key without echoing it.
// It tries `gum input --password` first; if gum is not on PATH, it reads from /dev/tty with term.ReadPassword.
func PromptSecretValue(ctx context.Context, key string) (string, error) {
	prompt := fmt.Sprintf("Valor para a variável de ambiente %s: ", key)
	s, err := promptSecretGum(ctx, prompt)
	if err == nil {
		return strings.TrimSpace(s), nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return promptSecretReadPassword(key)
	}
	return "", err
}

func promptSecretGum(ctx context.Context, prompt string) (string, error) {
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		return "", exec.ErrNotFound
	}
	args := []string{
		"input",
		"--password",
		"--prompt", prompt,
		"--placeholder", "Introduza o valor (oculto)",
		"--no-show-help",
	}
	cmd := exec.CommandContext(ctx, gumPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Env = ui.PrependGumThemeDefaults(os.Environ())
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func promptSecretReadPassword(key string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf(
			"comando 'gum' não encontrado e não foi possível abrir /dev/tty para ler o segredo de %q sem eco: %w",
			key,
			err,
		)
	}
	defer tty.Close()

	_, _ = fmt.Fprintf(os.Stderr, "Valor para %s (sem eco): ", key)
	line, err := term.ReadPassword(int(tty.Fd()))
	if err != nil {
		return "", fmt.Errorf("ler segredo para %q: %w", key, err)
	}
	_, _ = fmt.Fprintln(os.Stderr)
	return string(line), nil
}
