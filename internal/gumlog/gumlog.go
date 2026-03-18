// Package gumlog mirrors internal/helpers/shell/log.sh: gum log with --quiet / --verbose.
package gumlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"mb/internal/ui"
)

// Logger writes leveled messages via gum log (or a text fallback).
type Logger struct {
	Quiet   bool
	Verbose bool
	W       io.Writer // stderr by default
}

// New builds a logger. w is usually cmd.ErrOrStderr(); nil uses os.Stderr.
func New(quiet, verbose bool, w io.Writer) *Logger {
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

func sanitizeMsg(msg string) string {
	msg = strings.ReplaceAll(msg, "\r", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	return strings.TrimSpace(msg)
}

func (l *Logger) emit(ctx context.Context, level, msg string) error {
	if !l.shouldEmit(level) {
		return nil
	}
	msg = sanitizeMsg(msg)
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
