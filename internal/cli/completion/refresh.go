package completion

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/shared/system"
)

// TryRefreshInstalled regrava o bloco de completion no perfil por omissão do shell
// quando o ficheiro já contém BlockBegin. Ignora falhas de deteção (SHELL) ou ficheiro em falta.
func TryRefreshInstalled(
	ctx context.Context,
	root *cobra.Command,
	log *system.Logger,
	quiet bool,
) error {
	shell, err := DetectShell()
	if err != nil {
		return nil
	}
	rcPath, err := ProfilePath(shell, "")
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s := string(data)
	if !strings.Contains(s, BlockBegin) {
		return nil
	}
	var buf bytes.Buffer
	if err := WriteCompletionScript(root, shell, true, &buf); err != nil {
		return fmt.Errorf("gerar completion: %w", err)
	}
	marked := AppendMarkers(buf.String())
	newContent := MergeCompletionBlock(s, marked, BlockBegin, BlockEnd)
	if newContent == s {
		return nil
	}
	mode := os.FileMode(0o644)
	if fi, statErr := os.Stat(rcPath); statErr == nil {
		mode = fi.Mode().Perm()
	}
	if err := os.WriteFile(rcPath, []byte(newContent), mode); err != nil {
		return fmt.Errorf("gravar %s: %w", rcPath, err)
	}
	if log != nil && !quiet {
		_ = log.Info(ctx, "Autocompletar do shell atualizado em %s", rcPath)
	}
	return nil
}
