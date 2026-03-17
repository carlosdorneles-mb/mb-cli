package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// sanitizeTableCell removes characters that would break tab-separated columns in gum table.
func sanitizeTableCell(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// GumTable renders data with `gum table --print` (bordas arredondadas). Sem gum no PATH,
// imprime TSV simples.
func GumTable(ctx context.Context, headers []string, rows [][]string, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}
	if len(headers) == 0 {
		return nil
	}
	cleanHeaders := make([]string, len(headers))
	for i, h := range headers {
		cleanHeaders[i] = sanitizeTableCell(h)
	}
	cleanRows := make([][]string, len(rows))
	for i, row := range rows {
		cleanRows[i] = make([]string, len(row))
		for j, c := range row {
			if j < len(headers) {
				cleanRows[i][j] = sanitizeTableCell(c)
			}
		}
	}

	gumPath, err := exec.LookPath("gum")
	if err != nil {
		fmt.Fprintln(out, strings.Join(cleanHeaders, "\t"))
		for _, row := range cleanRows {
			for len(row) < len(cleanHeaders) {
				row = append(row, "")
			}
			fmt.Fprintln(out, strings.Join(row[:len(cleanHeaders)], "\t"))
		}
		return nil
	}

	args := []string{
		"table", "--print",
		"--separator", "\t",
		"--columns", strings.Join(cleanHeaders, ","),
		"--border", "rounded",
	}

	var input bytes.Buffer
	for _, row := range cleanRows {
		line := make([]string, len(cleanHeaders))
		for j := range cleanHeaders {
			if j < len(row) {
				line[j] = row[j]
			}
		}
		input.WriteString(strings.Join(line, "\t"))
		input.WriteString("\n")
	}

	cmd := exec.CommandContext(ctx, gumPath, args...)
	cmd.Stdin = &input
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func Input(ctx context.Context, prompt string) (string, error) {
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		return "", fmt.Errorf("gum not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, gumPath, "input", "--prompt", prompt+" ")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
