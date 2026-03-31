package completion

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// WriteCompletionScript gera o script de completion do mb para o shell indicado,
// com a mesma semântica dos subcomandos `mb completion <shell>`.
func WriteCompletionScript(root *cobra.Command, shell string, includeDesc bool, w io.Writer) error {
	switch shell {
	case ShellBash:
		return root.GenBashCompletionV2(w, includeDesc)
	case ShellZsh:
		if includeDesc {
			return root.GenZshCompletion(w)
		}
		return root.GenZshCompletionNoDesc(w)
	case ShellFish:
		return root.GenFishCompletion(w, includeDesc)
	case ShellPowerShell:
		if includeDesc {
			return root.GenPowerShellCompletionWithDesc(w)
		}
		return root.GenPowerShellCompletion(w)
	default:
		return fmt.Errorf("shell desconhecido: %s", shell)
	}
}
