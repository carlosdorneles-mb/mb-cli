package plugins

import "testing"

func TestPluginTypeFromEntrypoint(t *testing.T) {
	tests := []struct {
		entrypoint string
		want      string
	}{
		{"run.sh", "sh"},
		{"deploy.sh", "sh"},
		{"script.SH", "bin"},
		{"run", "bin"},
		{"bin/run", "bin"},
		{"run.bash", "bin"},
	}
	for _, tt := range tests {
		t.Run(tt.entrypoint, func(t *testing.T) {
			if got := PluginTypeFromEntrypoint(tt.entrypoint); got != tt.want {
				t.Errorf("PluginTypeFromEntrypoint(%q) = %q, want %q", tt.entrypoint, got, tt.want)
			}
		})
	}
}
