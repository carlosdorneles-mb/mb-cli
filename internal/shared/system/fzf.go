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

# Generate markdown with proper newlines
{
  echo "### $(echo "$ENTRY" | jq -r '.package // "N/A"')"
  echo ""
  echo "**Comando:** $(echo "$ENTRY" | jq -r '.command // "N/A"')"
  echo "**Descrição:** $(echo "$ENTRY" | jq -r '.description // "N/A"')"
  echo "**Versão:** $(echo "$ENTRY" | jq -r '.version // "N/A"')"
  echo "**Origem:** $(echo "$ENTRY" | jq -r '.origin // "N/A"')"
  echo "**URL:** $(echo "$ENTRY" | jq -r '.url // "N/A"')"

  REF=$(echo "$ENTRY" | jq -r '.ref // ""')
  if [ -n "$REF" ]; then
    REFTYPE=$(echo "$ENTRY" | jq -r '.refType // ""')
    if [ -n "$REFTYPE" ]; then
      echo "**Ref:** $REF ($REFTYPE)"
    else
      echo "**Ref:** $REF"
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
	selectedLine := strings.TrimSpace(output.String())
	if selectedLine == "" {
		return nil, nil
	}

	// Remove tab-separated index prefix and find matching row
	// selectedLine format is "INDEX\tFORMATTED_ROW"
	tabIdx := strings.Index(selectedLine, "\t")
	if tabIdx > 0 {
		indexStr := selectedLine[:tabIdx]
		var index int
		fmt.Sscanf(indexStr, "%d", &index)
		if index >= 0 && index < len(rows) {
			return rows[index], nil
		}
	}

	// Fallback: try matching by first column (package name)
	selectedParts := strings.Split(selectedLine, " │ ")
	if len(selectedParts) > 0 {
		firstCol := strings.TrimSpace(selectedParts[0])
		// Remove index prefix if still present
		if idx := strings.Index(firstCol, "\t"); idx > 0 {
			firstCol = firstCol[idx+1:]
		}
		for _, row := range rows {
			if len(row) > 0 && row[0] == firstCol {
				return row, nil
			}
		}
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

# Generate markdown with proper newlines
{
  echo "### $(echo "$ENTRY" | jq -r '.key // "N/A"')"
  echo ""
  echo "**Valor:** $(echo "$ENTRY" | jq -r '.displayValue // "N/A"')"
  echo "**Vault:** $(echo "$ENTRY" | jq -r '.vault // "N/A"')"
  echo "**Armazenamento:** $(echo "$ENTRY" | jq -r '.storage // "N/A"')"
  echo "**Arquivo:** $(echo "$ENTRY" | jq -r '.path // "N/A"')"

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
	selectedLine := strings.TrimSpace(output.String())
	if selectedLine == "" {
		return nil, nil
	}

	// Remove tab-separated index prefix and find matching row
	tabIdx := strings.Index(selectedLine, "\t")
	if tabIdx > 0 {
		indexStr := selectedLine[:tabIdx]
		var index int
		fmt.Sscanf(indexStr, "%d", &index)
		if index >= 0 && index < len(rows) {
			return rows[index], nil
		}
	}

	// Fallback: try matching by first column
	selectedParts := strings.Split(selectedLine, " │ ")
	if len(selectedParts) > 0 {
		firstCol := strings.TrimSpace(selectedParts[0])
		if idx := strings.Index(firstCol, "\t"); idx > 0 {
			firstCol = firstCol[idx+1:]
		}
		for _, row := range rows {
			if len(row) > 0 && row[0] == firstCol {
				return row, nil
			}
		}
	}

	return nil, nil
}
