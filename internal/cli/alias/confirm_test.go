package alias

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRequireConfirmOrYes_withYesSkipsPrompt(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	if err := requireConfirmOrYes(context.Background(), cmd, true, "nope"); err != nil {
		t.Fatal(err)
	}
}

func TestRequireConfirmOrYes_nonTTYRequiresYes(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader(""))
	cmd.SetErr(bytes.NewBuffer(nil))
	err := requireConfirmOrYes(context.Background(), cmd, false, "prompt")
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("expected error mentioning --yes, got %v", err)
	}
}
