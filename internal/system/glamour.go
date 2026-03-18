package system

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"charm.land/glamour/v2"
	"golang.org/x/term"
)

func RenderMarkdown(ctx context.Context, path string) error {
	_ = ctx // reserved for future use (e.g. cancellation)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := glamour.RenderWithEnvironmentConfig(string(content))
	if err != nil {
		return err
	}
	outBytes := []byte(out)
	if term.IsTerminal(int(os.Stdout.Fd())) {
		if pagerErr := runPager(outBytes); pagerErr == nil {
			return nil
		}
		// Pager indisponível (ex.: less não instalado); imprime direto
	}
	_, err = os.Stdout.Write(outBytes)
	return err
}

func runPager(content []byte) error {
	pagerCmd := os.Getenv("PAGER")
	if pagerCmd == "" {
		pagerCmd = "less -R"
	}
	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		return exec.ErrNotFound
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(string(content))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
