package gumlog

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLoggerQuietSkipsInfoWarnDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New(true, false, &buf)
	ctx := context.Background()
	_ = l.Debug(ctx, "d")
	_ = l.Info(ctx, "i")
	_ = l.Warn(ctx, "w")
	if buf.Len() != 0 {
		t.Errorf("quiet: expected no output, got %q", buf.String())
	}
	_ = l.Error(ctx, "e")
	if !strings.Contains(buf.String(), "e") && !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("quiet: expected error line, got %q", buf.String())
	}
}

func TestLoggerVerboseDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New(false, false, &buf)
	ctx := context.Background()
	_ = l.Debug(ctx, "hidden")
	if buf.Len() != 0 {
		t.Errorf("no verbose: debug should not write, got %q", buf.String())
	}
	buf.Reset()
	l2 := New(false, true, &buf)
	_ = l2.Debug(ctx, "shown")
	if !strings.Contains(buf.String(), "shown") {
		t.Errorf("verbose debug: got %q", buf.String())
	}
}

func TestLoggerInfoWhenNotQuiet(t *testing.T) {
	var buf bytes.Buffer
	l := New(false, false, &buf)
	ctx := context.Background()
	_ = l.Info(ctx, "hello %s", "world")
	if !strings.Contains(buf.String(), "hello") || !strings.Contains(buf.String(), "world") {
		t.Errorf("info: %q", buf.String())
	}
}

func TestSanitizeNewlines(t *testing.T) {
	var buf bytes.Buffer
	l := New(false, false, &buf)
	ctx := context.Background()
	_ = l.Info(ctx, "a\nb\nc")
	s := buf.String()
	if strings.Contains(s, "\n") && strings.Count(s, "\n") > 1 {
		t.Errorf("expected single line-ish output: %q", s)
	}
}
