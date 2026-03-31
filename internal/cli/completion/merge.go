package completion

import (
	"bytes"
	"strings"
)

// MergeCompletionBlock substitui um bloco existente entre beginMarker e endMarker,
// ou anexa o bloco no fim (com separação por newline). beginMarker e endMarker
// devem ser linhas completas (sem \n no interior).
func MergeCompletionBlock(existing, block string, beginMarker, endMarker string) string {
	block = strings.TrimRight(block, "\n")
	if block == "" {
		return existing
	}
	full := block
	if !strings.HasSuffix(full, "\n") {
		full += "\n"
	}

	lines := strings.Split(existing, "\n")
	beginIdx := -1
	endIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == beginMarker {
			beginIdx = i
			break
		}
	}
	if beginIdx >= 0 {
		for j := beginIdx + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == endMarker {
				endIdx = j
				break
			}
		}
	}

	var out []string
	if beginIdx >= 0 && endIdx >= 0 {
		out = append(out, lines[:beginIdx]...)
		out = append(out, strings.Split(strings.TrimRight(full, "\n"), "\n")...)
		out = append(out, lines[endIdx+1:]...)
	} else {
		switch trimmed := strings.TrimRight(existing, "\n"); trimmed {
		case "":
			out = []string{strings.TrimRight(full, "\n")}
		default:
			out = []string{trimmed, "", strings.TrimRight(full, "\n")}
		}
	}

	return strings.Join(out, "\n") + "\n"
}

// RemoveCompletionBlock remove as linhas entre beginMarker e endMarker inclusive.
// Se os marcadores não existirem em par, devolve existing inalterado.
func RemoveCompletionBlock(existing, beginMarker, endMarker string) string {
	lines := strings.Split(existing, "\n")
	beginIdx := -1
	endIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == beginMarker {
			beginIdx = i
			break
		}
	}
	if beginIdx < 0 {
		return existing
	}
	for j := beginIdx + 1; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == endMarker {
			endIdx = j
			break
		}
	}
	if endIdx < 0 {
		return existing
	}
	out := append([]string{}, lines[:beginIdx]...)
	out = append(out, lines[endIdx+1:]...)
	s := strings.Join(out, "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	return s + "\n"
}

// AppendMarkers wraps script body with begin/end marker lines.
func AppendMarkers(scriptBody string) string {
	var b bytes.Buffer
	b.WriteString(BlockBegin)
	b.WriteByte('\n')
	b.WriteString(strings.TrimRight(scriptBody, "\n"))
	b.WriteByte('\n')
	b.WriteString(BlockEnd)
	b.WriteByte('\n')
	return b.String()
}
