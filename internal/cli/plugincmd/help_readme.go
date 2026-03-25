package plugincmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mb/internal/shared/system"
)

const readmeFlagDesc = "Visualizar documentação do comando"

func setHelpFang(c *cobra.Command) {
	c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if root := cmd.Root(); root != nil {
			root.HelpFunc()(cmd, args)
		}
	})
}

func runReadmeWithGlow(path string) error {
	if path == "" {
		return errors.New("este comando não possui documentação (README) disponível")
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("documentação não encontrada para este comando")
		}
		return fmt.Errorf("não foi possível abrir a documentação: %w", err)
	}
	return system.RenderMarkdown(context.Background(), path)
}
