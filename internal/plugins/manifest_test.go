package plugins

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFlagsSpecListFormat(t *testing.T) {
	yamlList := `
- name: deploy
  description: Deploy
  entrypoint: deploy.sh
  commands:
    long: deploy
    short: d
- name: rollback
  description: Revert
  entrypoint: rollback.sh
  commands:
    long: rollback
    short: r
`
	var f FlagsSpec
	if err := yaml.Unmarshal([]byte(yamlList), &f); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if f.Len() != 2 {
		t.Errorf("Len() = %d, want 2", f.Len())
	}
	m := f.ToMap()
	if len(m) != 2 {
		t.Fatalf("ToMap() len = %d, want 2", len(m))
	}
	if d, ok := m["deploy"]; !ok || d.Entrypoint != "deploy.sh" || d.Short != "d" || d.Description != "Deploy" {
		t.Errorf("deploy entry: %+v", m["deploy"])
	}
	if r, ok := m["rollback"]; !ok || r.Description != "Revert" {
		t.Errorf("rollback entry: %+v", m["rollback"])
	}
}

func TestFlagsSpecMapFormatRejected(t *testing.T) {
	yamlMap := `
deploy:
  type: long
  short: d
  entrypoint: deploy.sh
rollback:
  type: long
  entrypoint: rollback.sh
`
	var f FlagsSpec
	err := yaml.Unmarshal([]byte(yamlMap), &f)
	if err == nil {
		t.Fatal("unmarshal of map format should fail")
	}
	if !strings.Contains(err.Error(), "list") {
		t.Errorf("error should mention list format: %v", err)
	}
}

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
