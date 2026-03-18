package system

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestConfirmFallbackYes(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var out bytes.Buffer
	ok, err := Confirm(context.Background(), "Remover?", strings.NewReader("y\n"), &out)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !strings.Contains(out.String(), "Remover?") {
		t.Errorf("prompt: %q", out.String())
	}
}

func TestConfirmFallbackNo(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	ok, err := Confirm(context.Background(), "X?", strings.NewReader("n\n"), io.Discard)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestConfirmFallbackEmptyLine(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	ok, err := Confirm(context.Background(), "X?", strings.NewReader("\n"), io.Discard)
	if err != nil || ok {
		t.Fatalf("empty: ok=%v err=%v", ok, err)
	}
}
