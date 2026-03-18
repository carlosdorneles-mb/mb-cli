package ui

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/x/term"
)

// ErrorHandlerPT is a fang.ErrorHandler that translates Cobra error messages to Portuguese
// and suggests "Use --help para uso." instead of "Try --help for usage."
func ErrorHandlerPT(w io.Writer, styles fang.Styles, err error) {
	if w, ok := w.(term.File); ok {
		if !term.IsTerminal(w.Fd()) {
			_, _ = fmt.Fprintln(w, translateError(err.Error()))
			return
		}
	}
	msg := translateError(err.Error())
	if msg != "" && !strings.HasSuffix(msg, ".") {
		msg += "."
	}
	_, _ = fmt.Fprintln(w, styles.ErrorHeader.SetString("ERRO").String())
	_, _ = fmt.Fprintln(w, styles.ErrorText.Render(msg))
	_, _ = fmt.Fprintln(w)
	if isUsageError(err.Error()) {
		_, _ = fmt.Fprintln(w, lipgloss.JoinHorizontal(
			lipgloss.Left,
			styles.ErrorText.UnsetWidth().Render("Use"),
			styles.Program.Flag.Render(" --help "),
			styles.ErrorText.UnsetWidth().UnsetMargins().UnsetTransform().Render("para detalhes."),
		))
		_, _ = fmt.Fprintln(w)
	}
}

func translateError(s string) string {
	switch {
	case strings.HasPrefix(s, "unknown command "):
		suffix := s[len("unknown command "):]
		suffix = strings.ReplaceAll(suffix, " for ", " para ")
		suffix = strings.ReplaceAll(suffix, "Did you mean this?", "Quis dizer:")
		return "Commando desconhecido " + suffix
	case strings.HasPrefix(s, "unknown flag: "):
		return "Flag desconhecida: " + s[len("unknown flag: "):]
	case strings.HasPrefix(s, "unknown shorthand flag: "):
		suffix := s[len("unknown shorthand flag: "):]
		suffix = strings.ReplaceAll(suffix, " in ", " em ")
		return "Flag curta desconhecida: " + suffix
	case strings.HasPrefix(s, "flag needs an argument: "):
		return "A flag precisa de um argumento: " + s[len("flag needs an argument: "):]
	case strings.HasPrefix(s, "invalid argument "):
		return "Argumento inválido " + s[len("invalid argument "):]
	case acceptArgCountRe.MatchString(s):
		return translateAcceptArgCount(s)
	default:
		return s
	}
}

// acceptArgCountRe matches Cobra "accepts X arg(s), received Y" (case-insensitive).
var acceptArgCountRe = regexp.MustCompile(`(?i)^accepts\s+(\d+)\s+arg\(s\),\s+received\s+(\d+)\s*$`)

func translateAcceptArgCount(s string) string {
	subs := acceptArgCountRe.FindStringSubmatch(s)
	if len(subs) != 3 {
		return s
	}
	return "Aceita " + subs[1] + " argumento(s), recebido(s) " + subs[2]
}

func isUsageError(errMsg string) bool {
	for _, prefix := range []string{
		"flag needs an argument:",
		"unknown flag:",
		"unknown shorthand flag:",
		"unknown command",
		"invalid argument",
	} {
		if strings.HasPrefix(errMsg, prefix) {
			return true
		}
	}
	if strings.HasPrefix(strings.ToLower(errMsg), "accepts ") &&
		acceptArgCountRe.MatchString(strings.TrimSpace(errMsg)) {
		return true
	}
	return false
}
