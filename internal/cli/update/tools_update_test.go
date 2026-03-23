package update

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestSaveRestoreRootArgs(t *testing.T) {
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(&cobra.Command{Use: "sub"})

	root.SetArgs([]string{"sub", "x"})
	prev, wasSet := saveRootArgs(root)
	if !wasSet {
		t.Fatal("expected wasSet true after SetArgs")
	}
	if !reflect.DeepEqual(prev, []string{"sub", "x"}) {
		t.Fatalf("saveRootArgs = %v", prev)
	}

	root.SetArgs([]string{"other"})
	restoreRootArgs(root, prev, wasSet)
	p2, _ := saveRootArgs(root)
	if !reflect.DeepEqual(p2, []string{"sub", "x"}) {
		t.Fatalf("after restore got %v", p2)
	}

	restoreRootArgs(root, nil, false)
	p3, was3 := saveRootArgs(root)
	if was3 {
		t.Fatalf("after restore wasSet=false, expected nil args, got %v", p3)
	}
}
