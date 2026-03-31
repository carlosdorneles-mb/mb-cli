package completion

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestWriteCompletionScript_smoke(t *testing.T) {
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(&cobra.Command{Use: "sub"})
	for _, shell := range []string{ShellBash, ShellZsh, ShellFish, ShellPowerShell} {
		var buf bytes.Buffer
		if err := WriteCompletionScript(root, shell, true, &buf); err != nil {
			t.Fatalf("%s: %v", shell, err)
		}
		if buf.Len() < 50 {
			t.Fatalf("%s: script too short: %d", shell, buf.Len())
		}
	}
}
