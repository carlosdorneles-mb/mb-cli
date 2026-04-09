package opcli

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsItemNotFound(t *testing.T) {
	t.Parallel()
	cases := []struct {
		msg   string
		want  bool
		label string
	}{
		{
			label: "op apostrophe form",
			msg:   `[ERROR] 2026/04/09 01:45:58 "mb-cli env / default" isn't an item. Specify the item with its UUID, name, or domain.`,
			want:  true,
		},
		{
			label: "uncontracted",
			msg:   `could not find item: "x" is not an item`,
			want:  true,
		},
		{
			label: "other",
			msg:   `op: permission denied`,
			want:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			err := errors.New(tc.msg)
			if got := isItemNotFound(err); got != tc.want {
				t.Fatalf("isItemNotFound(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestIsItemNotFoundWrapped(t *testing.T) {
	t.Parallel()
	inner := errors.New(`"x" isn't an item. Specify the item`)
	wrapped := fmt.Errorf("op item get %q: %w\nstderr", "title", inner)
	if !isItemNotFound(wrapped) {
		t.Fatal("expected wrapped error to be recognized as item not found")
	}
}
