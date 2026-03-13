package env

import (
	"strings"
	"testing"
)

func TestMergePrecedence(t *testing.T) {
	system := []string{"A=1", "B=1"}
	fileValues := map[string]string{"B": "2", "C": "2"}
	cliValues := map[string]string{"C": "3", "D": "3"}

	merged := Merge(system, fileValues, cliValues)
	got := AsMap(merged)

	if got["A"] != "1" {
		t.Fatalf("expected A to remain from system")
	}
	if got["B"] != "2" {
		t.Fatalf("expected B to be overridden by file")
	}
	if got["C"] != "3" {
		t.Fatalf("expected C to be overridden by cli")
	}
	if got["D"] != "3" {
		t.Fatalf("expected D from cli")
	}
}

func TestParseInlinePairs(t *testing.T) {
	values, err := ParseInlinePairs([]string{"K=V", "EMPTY=", "PATH=/tmp/a=b"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if values["K"] != "V" || values["EMPTY"] != "" || values["PATH"] != "/tmp/a=b" {
		t.Fatalf("parsed values mismatch: %#v", values)
	}
}

func TestParseInlinePairsRejectsMissingSeparator(t *testing.T) {
	_, err := ParseInlinePairs([]string{"INVALID"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "KEY=VALUE") {
		t.Fatalf("unexpected error: %v", err)
	}
}
