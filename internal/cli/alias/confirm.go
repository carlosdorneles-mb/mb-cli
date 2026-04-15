package alias

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/shared/system"
)

// ErrAliasOpCancelled is returned when the user declines a confirmation prompt.
var ErrAliasOpCancelled = errors.New("operação cancelada pelo usuário")

// requireConfirmOrYes enforces an interactive confirmation unless --yes was passed.
// In non-interactive environments without --yes, returns an error.
func requireConfirmOrYes(ctx context.Context, cmd *cobra.Command, yes bool, prompt string) error {
	if yes {
		return nil
	}
	in := cmd.InOrStdin()
	f, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return fmt.Errorf(
			"é necessário confirmar esta alteração num terminal interativo ou usar --yes",
		)
	}
	okConfirm, err := system.Confirm(ctx, prompt, in, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if !okConfirm {
		return ErrAliasOpCancelled
	}
	return nil
}
