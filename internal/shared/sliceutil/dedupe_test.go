package sliceutil

import (
	"reflect"
	"testing"
)

func TestDedupeStringsPreserveOrder(t *testing.T) {
	t.Parallel()
	in := []string{"a", "b", "a", "c", "b", "a"}
	got := DedupeStringsPreserveOrder(in)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got := DedupeStringsPreserveOrder(nil); len(got) != 0 {
		t.Fatalf("nil: got %v", got)
	}
}
