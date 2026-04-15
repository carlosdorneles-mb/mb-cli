package system

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// PluginEntry represents a plugin entry for fzf preview and data display.
type PluginEntry struct {
	Package         string `json:"package"`
	Command         string `json:"command"`
	Description     string `json:"description"`
	Version         string `json:"version"`
	Origin          string `json:"origin"`
	URL             string `json:"url"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Ref             string `json:"ref,omitempty"`
	RefType         string `json:"refType,omitempty"`
}

// EnvEntry represents an environment variable entry for fzf preview.
type EnvEntry struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DisplayValue string `json:"displayValue"`
	Vault        string `json:"vault"`
	Storage      string `json:"storage"`
	IsSecret     bool   `json:"isSecret"`
	Path         string `json:"path"`
}

// AliasEntry represents an alias row for fzf preview (command is joined argv for jq in preview.sh).
type AliasEntry struct {
	Name      string `json:"name"`
	EnvVault  string `json:"envVault"`
	Command   string `json:"command"`
	Source    string `json:"source,omitempty"`
	MbcliPath string `json:"mbcliPath,omitempty"`
}

// FzfTable displays data in fzf table mode with headers and returns the selected row.
// If not a TTY or fzf not found, falls back to GumTable.
func FzfTable(
	ctx context.Context,
	headers []string,
	rows [][]string,
	out io.Writer,
) (selectedRow []string, err error) {
	// Always render something if there are no rows
	if len(rows) == 0 {
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		// Not a terminal, use GumTable
		return nil, GumTable(ctx, headers, rows, out)
	}

	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		// fzf not found, use GumTable
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Calculate column widths for proper alignment
	widths := calculateColumnWidths(headers, rows)

	// Create formatted header text
	headerText := createHeaderText(headers, widths)

	// Format rows with fixed widths (no header in stdin)
	var input bytes.Buffer
	for _, row := range rows {
		input.WriteString(formatRowFixedWidth(row, widths))
		input.WriteString("\n")
	}

	// Build fzf command with header as fixed text
	args := []string{
		"--prompt", "Filtrar> ",
		"--layout", "reverse",
		"--height", "~80%",
		"--cycle",
		"--header", headerText,
		"--bind", "enter:accept",
		"--bind", "esc:cancel",
		"--bind", "ctrl-c:cancel",
	}

	// Capture fzf output
	var output bytes.Buffer
	cmd := exec.CommandContext(ctx, fzfPath, args...)
	cmd.Stdin = &input
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				// User cancelled with ESC - just exit, no fallback
				return nil, nil
			}
		}
		// Other error, fall back to GumTable
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Parse selected line
	selectedLine := strings.TrimSpace(output.String())
	if selectedLine == "" {
		return nil, nil
	}

	// Find the matching row using fixed-width formatting
	for _, row := range rows {
		formatted := formatRowFixedWidth(row, widths)
		if formatted == selectedLine {
			return row, nil
		}
	}

	// Fallback: try matching by first column (package name)
	selectedParts := strings.Split(selectedLine, " │ ")
	if len(selectedParts) > 0 {
		selectedPkg := strings.TrimSpace(selectedParts[0])
		for _, row := range rows {
			if len(row) > 0 && row[0] == selectedPkg {
				return row, nil
			}
		}
	}

	// If no exact match, return nil
	return nil, nil
}

// calculateColumnWidths calculates the maximum width for each column based on headers and data.
func calculateColumnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))

	// Headers
	for i, h := range headers {
		widths[i] = len(h)
	}

	// Data rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Add padding (2 spaces on each side)
	for i := range widths {
		widths[i] += 4
	}

	return widths
}

// formatRowFixedWidth formats a row with fixed column widths for proper alignment.
func formatRowFixedWidth(row []string, widths []int) string {
	var sb strings.Builder
	for i := range widths {
		if i > 0 {
			sb.WriteString(" │ ")
		}
		cell := ""
		if i < len(row) {
			cell = sanitizeTableCell(row[i])
		}
		// Pad right to fixed width
		fmt.Fprintf(&sb, "%-*s", widths[i], cell)
	}
	return sb.String()
}

// createHeaderText creates a formatted header line with fixed column widths.
func createHeaderText(headers []string, widths []int) string {
	var sb strings.Builder
	for i := range widths {
		if i > 0 {
			sb.WriteString(" │ ")
		}
		cell := ""
		if i < len(headers) {
			cell = headers[i]
		}
		// Pad right to fixed width
		fmt.Fprintf(&sb, "%-*s", widths[i], cell)
	}
	return sb.String()
}

// parseFzfIndexedSelection parses fzf stdout "N\t<formatted row>" where N is a 1-based index into rows.
func parseFzfIndexedSelection(selectedLine string, rows [][]string) ([]string, bool) {
	selectedLine = strings.TrimSpace(selectedLine)
	if selectedLine == "" {
		return nil, false
	}
	tabIdx := strings.Index(selectedLine, "\t")
	if tabIdx > 0 {
		var n int
		if _, err := fmt.Sscanf(selectedLine[:tabIdx], "%d", &n); err == nil {
			if n >= 1 && n <= len(rows) {
				return rows[n-1], true
			}
		}
	}
	selectedParts := strings.Split(selectedLine, " │ ")
	if len(selectedParts) > 0 {
		firstCol := strings.TrimSpace(selectedParts[0])
		if idx := strings.Index(firstCol, "\t"); idx > 0 {
			firstCol = firstCol[idx+1:]
		}
		for _, row := range rows {
			if len(row) > 0 && row[0] == firstCol {
				return row, true
			}
		}
	}
	return nil, false
}

// FzfTableWithPreview displays data in fzf with a preview panel using gum format.
// The previewFn is called to get preview text, but we use gum format for better rendering.
func FzfTableWithPreview(
	ctx context.Context,
	headers []string,
	rows [][]string,
	out io.Writer,
	entries []PluginEntry,
) (selectedRow []string, err error) {
	if len(rows) == 0 {
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, GumTable(ctx, headers, rows, out)
	}

	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Check if gum is available for preview
	_, gumErr := exec.LookPath("gum")
	if gumErr != nil {
		// No gum, fall back to regular FzfTable
		return FzfTable(ctx, headers, rows, out)
	}
	if _, err := exec.LookPath("jq"); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	// Calculate column widths for proper alignment
	widths := calculateColumnWidths(headers, rows)

	// Create formatted header text
	headerText := createHeaderText(headers, widths)

	// Format rows with tab-separated index (hidden from display)
	// Índices começam em 1 para o script preview
	var input bytes.Buffer
	for i, row := range rows {
		fmt.Fprintf(&input, "%d\t%s", i+1, formatRowFixedWidth(row, widths))
		input.WriteString("\n")
	}

	// Create temp directory for preview script and data
	tmpDir, err := os.MkdirTemp("", "mb-fzf-*")
	if err != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	defer os.RemoveAll(tmpDir)

	// Write entries to JSON file
	entriesJSON, err := json.Marshal(entries)
	if err != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	entriesFile := filepath.Join(tmpDir, "entries.json")
	if err := os.WriteFile(entriesFile, entriesJSON, 0644); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	// Create preview script
	previewScript := filepath.Join(tmpDir, "preview.sh")
	scriptContent := fmt.Sprintf(`#!/bin/bash
NUMINDEX="$1"
if [ -z "$NUMINDEX" ]; then
  echo "**Nenhum plugin selecionado**" | gum format
  exit 0
fi

# Converter índice 1-based para 0-based
INDEX=$(($NUMINDEX - 1))
if [ "$INDEX" -lt 0 ]; then
  echo "**Nenhum plugin selecionado**" | gum format
  exit 0
fi

ENTRIES_FILE="%s"
ENTRY=$(cat "$ENTRIES_FILE" | jq ".[$INDEX]")
if [ -z "$ENTRY" ] || [ "$ENTRY" = "null" ]; then
  echo "**Erro ao ler dados do plugin**" | gum format
  exit 0
fi

# Generate markdown with proper newlines (jq via vars + printf: avoids $ re-expansion inside "echo ... $(jq)" )
PKG=$(echo "$ENTRY" | jq -r '.package // "N/A"')
CMD=$(echo "$ENTRY" | jq -r '.command // "N/A"')
DESC=$(echo "$ENTRY" | jq -r '.description // "N/A"')
VER=$(echo "$ENTRY" | jq -r '.version // "N/A"')
ORIG=$(echo "$ENTRY" | jq -r '.origin // "N/A"')
URL=$(echo "$ENTRY" | jq -r '.url // "N/A"')
{
  printf '%%s\n' "### ${PKG}"
  echo ""
  printf '%%s\n' "**Comando:** ${CMD}"
  printf '%%s\n' "**Descrição:** ${DESC}"
  printf '%%s\n' "**Versão:** ${VER}"
  printf '%%s\n' "**Origem:** ${ORIG}"
  printf '%%s\n' "**URL:** ${URL}"

  REF=$(echo "$ENTRY" | jq -r '.ref // ""')
  if [ -n "$REF" ]; then
    REFTYPE=$(echo "$ENTRY" | jq -r '.refType // ""')
    if [ -n "$REFTYPE" ]; then
      printf '%%s\n' "**Ref:** ${REF} (${REFTYPE})"
    else
      printf '%%s\n' "**Ref:** ${REF}"
    fi
  fi
} | gum format
`, entriesFile)

	if err := os.WriteFile(previewScript, []byte(scriptContent), 0755); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	// Build fzf command with preview
	args := []string{
		"--prompt", "Filtrar> ",
		"--layout", "reverse",
		"--height", "~80%",
		"--cycle",
		"--header", headerText,
		"--delimiter", "\t",
		"--with-nth", "2..",
		"--preview", previewScript + " {1}",
		"--preview-window", "right:50%",
		"--bind", "enter:accept",
		"--bind", "esc:cancel",
		"--bind", "ctrl-c:cancel",
	}

	// Capture fzf output
	var output bytes.Buffer
	cmd := exec.CommandContext(ctx, fzfPath, args...)
	cmd.Stdin = &input
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				// User cancelled with ESC - just exit, no fallback
				return nil, nil
			}
		}
		// Other error, fall back to GumTable
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Parse selected line
	if row, ok := parseFzfIndexedSelection(output.String(), rows); ok {
		return row, nil
	}
	return nil, nil
}

// FzfTableWithPreviewForEnv displays env variables in fzf with a preview panel using gum format.
func FzfTableWithPreviewForEnv(
	ctx context.Context,
	headers []string,
	rows [][]string,
	out io.Writer,
	entries []EnvEntry,
) (selectedRow []string, err error) {
	if len(rows) == 0 {
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, GumTable(ctx, headers, rows, out)
	}

	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Check if gum is available for preview
	_, gumErr := exec.LookPath("gum")
	if gumErr != nil {
		// No gum, fall back to regular FzfTable
		return FzfTable(ctx, headers, rows, out)
	}
	if _, err := exec.LookPath("jq"); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	// Calculate column widths for proper alignment
	widths := calculateColumnWidths(headers, rows)

	// Create formatted header text
	headerText := createHeaderText(headers, widths)

	// Format rows with tab-separated index (hidden from display)
	// Índices começam em 1 para o script preview
	var input bytes.Buffer
	for i, row := range rows {
		fmt.Fprintf(&input, "%d\t%s", i+1, formatRowFixedWidth(row, widths))
		input.WriteString("\n")
	}

	// Create temp directory for preview script and data
	tmpDir, err := os.MkdirTemp("", "mb-fzf-*")
	if err != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	defer os.RemoveAll(tmpDir)

	// Write entries to JSON file
	entriesJSON, err := json.Marshal(entries)
	if err != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	entriesFile := filepath.Join(tmpDir, "entries.json")
	if err := os.WriteFile(entriesFile, entriesJSON, 0644); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	// Create preview script
	previewScript := filepath.Join(tmpDir, "preview.sh")
	scriptContent := fmt.Sprintf(`#!/bin/bash
NUMINDEX="$1"
if [ -z "$NUMINDEX" ]; then
  echo "**Nenhuma variável selecionada**" | gum format
  exit 0
fi

# Converter índice 1-based para 0-based
INDEX=$(($NUMINDEX - 1))
if [ "$INDEX" -lt 0 ]; then
  echo "**Nenhuma variável selecionada**" | gum format
  exit 0
fi

ENTRIES_FILE="%s"
ENTRY=$(cat "$ENTRIES_FILE" | jq ".[$INDEX]")
if [ -z "$ENTRY" ] || [ "$ENTRY" = "null" ]; then
  echo "**Erro ao ler dados da variável**" | gum format
  exit 0
fi

# Generate markdown with proper newlines (jq via vars + printf: avoids $ re-expansion inside "echo ... $(jq)" )
KEY=$(echo "$ENTRY" | jq -r '.key // "N/A"')
DISPLAY=$(echo "$ENTRY" | jq -r '.displayValue // "N/A"')
VAULT_NAME=$(echo "$ENTRY" | jq -r '.vault // "N/A"')
STORAGE=$(echo "$ENTRY" | jq -r '.storage // "N/A"')
ENV_PATH=$(echo "$ENTRY" | jq -r '.path // "N/A"')
{
  printf '%%s\n' "### ${KEY}"
  echo ""
  printf '%%s\n' "**Valor:** ${DISPLAY}"
  printf '%%s\n' "**Vault:** ${VAULT_NAME}"
  printf '%%s\n' "**Armazenamento:** ${STORAGE}"
  printf '%%s\n' "**Arquivo:** ${ENV_PATH}"

  IS_SECRET=$(echo "$ENTRY" | jq -r '.isSecret // false')
  if [ "$IS_SECRET" = "true" ]; then
    echo ""
    echo "Use --show-secrets para ver o valor real"
  fi
} | gum format
`, entriesFile)

	if err := os.WriteFile(previewScript, []byte(scriptContent), 0755); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	// Build fzf command with preview
	args := []string{
		"--prompt", "Filtrar> ",
		"--layout", "reverse",
		"--height", "~80%",
		"--cycle",
		"--header", headerText,
		"--delimiter", "\t",
		"--with-nth", "2..",
		"--preview", previewScript + " {1}",
		"--preview-window", "right:50%",
		"--bind", "enter:accept",
		"--bind", "esc:cancel",
		"--bind", "ctrl-c:cancel",
	}

	// Capture fzf output
	var output bytes.Buffer
	cmd := exec.CommandContext(ctx, fzfPath, args...)
	cmd.Stdin = &input
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				// User cancelled with ESC - just exit, no fallback
				return nil, nil
			}
		}
		// Other error, fall back to GumTable
		return nil, GumTable(ctx, headers, rows, out)
	}

	// Parse selected line
	if row, ok := parseFzfIndexedSelection(output.String(), rows); ok {
		return row, nil
	}
	return nil, nil
}

// FzfTableWithPreviewForAliases displays aliases in fzf with a preview panel (vault for mb run + command).
func FzfTableWithPreviewForAliases(
	ctx context.Context,
	headers []string,
	rows [][]string,
	out io.Writer,
	entries []AliasEntry,
) (selectedRow []string, err error) {
	if len(rows) == 0 {
		return nil, GumTable(ctx, headers, rows, out)
	}

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, GumTable(ctx, headers, rows, out)
	}

	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, GumTable(ctx, headers, rows, out)
	}

	_, gumErr := exec.LookPath("gum")
	if gumErr != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	if _, err := exec.LookPath("jq"); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	widths := calculateColumnWidths(headers, rows)
	headerText := createHeaderText(headers, widths)

	var input bytes.Buffer
	for i, row := range rows {
		fmt.Fprintf(&input, "%d\t%s", i+1, formatRowFixedWidth(row, widths))
		input.WriteString("\n")
	}

	tmpDir, err := os.MkdirTemp("", "mb-fzf-*")
	if err != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	defer os.RemoveAll(tmpDir)

	entriesJSON, err := json.Marshal(entries)
	if err != nil {
		return FzfTable(ctx, headers, rows, out)
	}
	entriesFile := filepath.Join(tmpDir, "entries.json")
	if err := os.WriteFile(entriesFile, entriesJSON, 0644); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	previewScript := filepath.Join(tmpDir, "preview.sh")
	scriptContent := fmt.Sprintf(`#!/bin/bash
NUMINDEX="$1"
if [ -z "$NUMINDEX" ]; then
  echo "**Nenhum alias selecionado**" | gum format
  exit 0
fi

INDEX=$(($NUMINDEX - 1))
if [ "$INDEX" -lt 0 ]; then
  echo "**Nenhum alias selecionado**" | gum format
  exit 0
fi

ENTRIES_FILE="%s"
ENTRY=$(cat "$ENTRIES_FILE" | jq ".[$INDEX]")
if [ -z "$ENTRY" ] || [ "$ENTRY" = "null" ]; then
  echo "**Erro ao ler dados do alias**" | gum format
  exit 0
fi

VAULT_RAW=$(echo "$ENTRY" | jq -r '.envVault // ""')
if [ -z "$VAULT_RAW" ]; then
  VAULT_LABEL="(nenhum)"
else
  VAULT_LABEL="$VAULT_RAW"
fi
NAME=$(echo "$ENTRY" | jq -r '.name // "N/A"')
CMD=$(echo "$ENTRY" | jq -r '.command // "N/A"')
SOURCE=$(echo "$ENTRY" | jq -r '.source // ""')
MBCLI=$(echo "$ENTRY" | jq -r '.mbcliPath // ""')
{
  printf '%%s\n' "### ${NAME}"
  echo ""
  printf '%%s\n' "**Vault (mb run):** ${VAULT_LABEL}"
  printf '%%s\n' "**Comando:** ${CMD}"
  if [ -n "$SOURCE" ]; then
    printf '%%s\n' "**Origem:** ${SOURCE}"
  fi
  if [ -n "$MBCLI" ]; then
    printf '%%s\n' "**Arquivo mbcli.yaml:** ${MBCLI}"
  fi
} | gum format
`, entriesFile)

	if err := os.WriteFile(previewScript, []byte(scriptContent), 0755); err != nil {
		return FzfTable(ctx, headers, rows, out)
	}

	args := []string{
		"--prompt", "Filtrar> ",
		"--layout", "reverse",
		"--height", "~80%",
		"--cycle",
		"--header", headerText,
		"--delimiter", "\t",
		"--with-nth", "2..",
		"--preview", previewScript + " {1}",
		"--preview-window", "right:50%",
		"--bind", "enter:accept",
		"--bind", "esc:cancel",
		"--bind", "ctrl-c:cancel",
	}

	var output bytes.Buffer
	cmd := exec.CommandContext(ctx, fzfPath, args...)
	cmd.Stdin = &input
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				return nil, nil
			}
		}
		return nil, GumTable(ctx, headers, rows, out)
	}

	if row, ok := parseFzfIndexedSelection(output.String(), rows); ok {
		return row, nil
	}
	return nil, nil
}
