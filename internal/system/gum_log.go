package system

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"mb/internal/shared/ui"
)

// Logger writes leveled messages via gum log (or a text fallback). Mirrors log.sh quiet/verbose rules.
type Logger struct {
	Quiet   bool
	Verbose bool
	W       io.Writer
}

// NewLogger builds a logger. w is usually cmd.ErrOrStderr(); nil uses os.Stderr.
func NewLogger(quiet, verbose bool, w io.Writer) *Logger {
	return &Logger{Quiet: quiet, Verbose: verbose, W: w}
}

func (l *Logger) out() io.Writer {
	if l.W != nil {
		return l.W
	}
	return os.Stderr
}

func (l *Logger) shouldEmit(level string) bool {
	if l.Quiet && level != "error" && level != "fatal" {
		return false
	}
	if !l.Verbose && level == "debug" {
		return false
	}
	return true
}

func sanitizeGumLogMsg(msg string) string {
	msg = strings.ReplaceAll(msg, "\r", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	return strings.TrimSpace(msg)
}

func (l *Logger) emit(ctx context.Context, level, msg string) error {
	if !l.shouldEmit(level) {
		return nil
	}
	msg = sanitizeGumLogMsg(msg)
	if msg == "" {
		return nil
	}
	w := l.out()
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		_, err = fmt.Fprintf(w, "[%s] %s\n", strings.ToUpper(level), msg)
		return err
	}
	cmd := exec.CommandContext(ctx, gumPath, "log", "-l", level, msg)
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Env = ui.PrependGumThemeDefaults(os.Environ())
	return cmd.Run()
}

func (l *Logger) Debug(ctx context.Context, format string, a ...interface{}) error {
	return l.emit(ctx, "debug", fmt.Sprintf(format, a...))
}

func (l *Logger) Info(ctx context.Context, format string, a ...interface{}) error {
	return l.emit(ctx, "info", fmt.Sprintf(format, a...))
}

func (l *Logger) Warn(ctx context.Context, format string, a ...interface{}) error {
	return l.emit(ctx, "warn", fmt.Sprintf(format, a...))
}

func (l *Logger) Error(ctx context.Context, format string, a ...interface{}) error {
	return l.emit(ctx, "error", fmt.Sprintf(format, a...))
}
