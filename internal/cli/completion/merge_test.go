package completion

import (
	"strings"
	"testing"
)

func TestMergeCompletionBlock_appendEmpty(t *testing.T) {
	block := AppendMarkers("line1\nline2")
	got := MergeCompletionBlock("", block, BlockBegin, BlockEnd)
	hasAll := strings.Contains(got, BlockBegin)
	hasAll = hasAll && strings.Contains(got, "line1")
	hasAll = hasAll && strings.Contains(got, BlockEnd)
	if !hasAll {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestMergeCompletionBlock_replaceIdempotent(t *testing.T) {
	first := AppendMarkers("v1")
	base := "before\n"
	merged := MergeCompletionBlock(base, first, BlockBegin, BlockEnd)
	second := AppendMarkers("v2")
	merged2 := MergeCompletionBlock(merged, second, BlockBegin, BlockEnd)
	if strings.Count(merged2, BlockBegin) != 1 {
		t.Fatalf("expected single BEGIN, got:\n%s", merged2)
	}
	if !strings.Contains(merged2, "v2") || strings.Contains(merged2, "v1") {
		t.Fatalf("expected v2 only, got:\n%s", merged2)
	}
	if !strings.HasPrefix(merged2, "before\n") {
		t.Fatalf("expected prefix preserved, got:\n%s", merged2)
	}
}

func TestMergeCompletionBlock_preserveAfter(t *testing.T) {
	base := "top\n" + AppendMarkers("mid") + "bottom\n"
	newBlock := AppendMarkers("newmid")
	got := MergeCompletionBlock(base, newBlock, BlockBegin, BlockEnd)
	if !strings.Contains(got, "top") {
		t.Fatalf("missing top:\n%s", got)
	}
	if !strings.Contains(got, "bottom") {
		t.Fatalf("missing bottom:\n%s", got)
	}
	if !strings.Contains(got, "newmid") {
		t.Fatalf("missing newmid:\n%s", got)
	}
	if strings.Contains(got, "\nmid\n") {
		t.Fatalf("old inner line should be gone:\n%s", got)
	}
}

func TestRemoveCompletionBlock_removesMarkedRegion(t *testing.T) {
	base := "keep\n" + AppendMarkers("body") + "tail\n"
	got := RemoveCompletionBlock(base, BlockBegin, BlockEnd)
	if strings.Contains(got, BlockBegin) || strings.Contains(got, "body") {
		t.Fatalf("block should be gone:\n%s", got)
	}
	if !strings.Contains(got, "keep") || !strings.Contains(got, "tail") {
		t.Fatalf("expected keep and tail:\n%s", got)
	}
}

func TestRemoveCompletionBlock_noMarkersUnchanged(t *testing.T) {
	base := "plain\n"
	if RemoveCompletionBlock(base, BlockBegin, BlockEnd) != base {
		t.Fatal("expected unchanged")
	}
}

func TestRemoveCompletionBlock_onlyBeginUnchanged(t *testing.T) {
	base := "x\n" + BlockBegin + "\norphan\n"
	if RemoveCompletionBlock(base, BlockBegin, BlockEnd) != base {
		t.Fatal("expected unchanged without END")
	}
}
