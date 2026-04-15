package alias

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cli/completion"
	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

// EnsureProfileOptions controls shell profile integration for user aliases.
type EnsureProfileOptions struct {
	ShellFlag string
}

// skipAliasProfileForHelp returns true when os.Args indicate a help-only invocation
// under `mb alias` (e.g. mb alias --help, mb alias set --help).
func skipAliasProfileForHelp(args []string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] != "alias" {
			continue
		}
		for j := i + 1; j < len(args); j++ {
			if args[j] == "-h" || args[j] == "--help" {
				return true
			}
		}
		return false
	}
	return false
}

// extractAliasBlockInner returns the raw body between BlockBegin and BlockEnd, or ok=false.
func extractAliasBlockInner(content string) (inner string, ok bool) {
	lines := strings.Split(content, "\n")
	beginIdx := -1
	endIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == BlockBegin {
			beginIdx = i
			break
		}
	}
	if beginIdx < 0 {
		return "", false
	}
	for j := beginIdx + 1; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == BlockEnd {
			endIdx = j
			break
		}
	}
	if endIdx < 0 {
		return "", false
	}
	innerLines := lines[beginIdx+1 : endIdx]
	return strings.Join(innerLines, "\n"), true
}

func nonEmptyTrimmedLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// ProfileHasExpectedAliasBlock reports whether the profile already contains our block
// with exactly one non-empty inner line matching wantLine.
func ProfileHasExpectedAliasBlock(profileContent, wantLine string) bool {
	inner, ok := extractAliasBlockInner(profileContent)
	if !ok {
		return false
	}
	lines := nonEmptyTrimmedLines(inner)
	want := strings.TrimSpace(wantLine)
	return len(lines) == 1 && lines[0] == want
}

func appendMarkers(body string) string {
	body = strings.TrimRight(body, "\n")
	var b strings.Builder
	b.WriteString(BlockBegin)
	b.WriteByte('\n')
	b.WriteString(body)
	b.WriteByte('\n')
	b.WriteString(BlockEnd)
	b.WriteByte('\n')
	return b.String()
}

// ensureShellProfileForAliases writes generated scripts and ensures the shell profile
// contains the mb-cli aliases block (automatic write to the shell default profile path).
func ensureShellProfileForAliases(
	d deps.Dependencies,
	cmd *cobra.Command,
	opts EnsureProfileOptions,
) error {
	cfgDir := d.Runtime.ConfigDir
	if err := alib.WriteShellScripts(cfgDir); err != nil {
		return fmt.Errorf("gerar scripts de shell: %w", err)
	}

	shell := opts.ShellFlag
	var err error
	if shell == "" {
		shell, err = completion.DetectShell()
		if err != nil {
			return err
		}
	} else {
		shell, err = completion.NormalizeShellName(shell)
		if err != nil {
			return err
		}
	}

	line, err := profileSourceLine(cfgDir, shell)
	if err != nil {
		return err
	}
	marked := appendMarkers(line)

	rcPath, err := completion.ProfilePath(shell, "")
	if err != nil {
		return err
	}

	existingBytes, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ler %s: %w", rcPath, err)
	}
	existing := string(existingBytes)

	if ProfileHasExpectedAliasBlock(existing, line) {
		return nil
	}

	newContent := completion.MergeCompletionBlock(existing, marked, BlockBegin, BlockEnd)

	firstRegistration := !strings.Contains(existing, BlockBegin)

	if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
		return fmt.Errorf("criar diretório: %w", err)
	}

	mode := os.FileMode(0o644)
	if fi, statErr := os.Stat(rcPath); statErr == nil {
		mode = fi.Mode().Perm()
	}
	if err := os.WriteFile(rcPath, []byte(newContent), mode); err != nil {
		return fmt.Errorf("salvar %s: %w", rcPath, err)
	}

	if firstRegistration {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
		msg := "Integração com o shell: foi adicionado ao perfil «%s» o carregamento automático dos aliases do MB CLI. " +
			"Abra um novo terminal ou execute source nesse arquivo para passar a usar os nomes dos aliases."
		_ = log.Info(ctx, msg, rcPath)
	}
	return nil
}
