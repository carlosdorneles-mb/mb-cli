package system

import (
	"context"
	"os"

	"github.com/charmbracelet/glamour"
)

func RenderMarkdown(ctx context.Context, path string) error {
	_ = ctx // reserved for future use (e.g. cancellation)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		return err
	}
	out, err := r.RenderBytes(content)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
