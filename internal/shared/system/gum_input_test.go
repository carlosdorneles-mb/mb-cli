package system

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestInput_gumNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := Input(context.Background(), "name?")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, exec.ErrNotFound) && !strings.Contains(err.Error(), "gum") {
		t.Errorf("err: %v", err)
	}
}
