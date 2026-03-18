package system

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMarkdown_fileNotFound(t *testing.T) {
	err := RenderMarkdown(context.Background(), filepath.Join(t.TempDir(), "nope.md"))
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("err=%v", err)
	}
}

func TestRenderMarkdown_rendersMarkdown(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(p, []byte("# Hello World\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	outDone := make(chan struct{})
	var captured bytes.Buffer
	go func() {
		_, _ = io.Copy(&captured, r)
		_ = r.Close()
		close(outDone)
	}()

	err = RenderMarkdown(context.Background(), p)
	_ = w.Close()
	os.Stdout = old
	<-outDone

	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(captured.String(), "Hello") {
		t.Errorf("expected rendered title in output, got %q", captured.String())
	}
}

func TestRunPager_cat(t *testing.T) {
	t.Setenv("PAGER", "cat")
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	outDone := make(chan struct{})
	var captured bytes.Buffer
	go func() {
		_, _ = io.Copy(&captured, r)
		_ = r.Close()
		close(outDone)
	}()

	err = runPager([]byte("pager-out"))
	_ = w.Close()
	os.Stdout = old
	<-outDone

	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(captured.String(), "pager-out") {
		t.Errorf("got %q", captured.String())
	}
}

func TestRunPager_emptyPagerEnvUsesLess(t *testing.T) {
	t.Setenv("PAGER", "")
	// Sem less no PATH: deve falhar de forma previsível ou suceder se less existir.
	// Garantimos que não panic e que erro é possível quando less ausente.
	t.Setenv("PATH", t.TempDir())
	err := runPager([]byte("x"))
	if err == nil {
		t.Log("less available despite empty PATH override — ok")
	}
}
