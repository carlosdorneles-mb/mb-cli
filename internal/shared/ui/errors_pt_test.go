package ui

import (
	"testing"
)

func TestTranslateError_AcceptArgCount(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"accepts 1 arg(s), received 0", "Aceita 1 argumento(s), recebido(s) 0"},
		{"Accepts 1 arg(s), received 0", "Aceita 1 argumento(s), recebido(s) 0"},
		{"accepts 2 arg(s), received 1", "Aceita 2 argumento(s), recebido(s) 1"},
	}
	for _, tt := range tests {
		got := translateError(tt.in)
		if got != tt.want {
			t.Errorf("translateError(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsUsageError_AcceptArgCount(t *testing.T) {
	msgs := []string{
		"accepts 1 arg(s), received 0",
		"Accepts 1 arg(s), received 0",
	}
	for _, msg := range msgs {
		if !isUsageError(msg) {
			t.Errorf("isUsageError(%q) = false, want true", msg)
		}
	}
}
