package plugins

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFlagsSpecListFormat(t *testing.T) {
	yamlList := "- name: deploy\n" +
		"  description: Deploy\n" +
		"  entrypoint: deploy.sh\n" +
		"  envs:\n" +
		"    - DEPLOY=true\n" +
		"    - REGION=us-east-1\n" +
		"  commands:\n" +
		"    long: deploy\n" +
		"    short: d\n" +
		"- name: rollback\n" +
		"  description: Revert\n" +
		"  entrypoint: rollback.sh\n" +
		"  commands:\n" +
		"    long: rollback\n" +
		"    short: r\n"
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
	if d, ok := m["deploy"]; !ok || d.Entrypoint != "deploy.sh" || d.Short != "d" ||
		d.Description != "Deploy" {
		t.Errorf("deploy entry: %+v", m["deploy"])
	}
	if got := m["deploy"].Envs; len(got) != 2 || got[0] != "DEPLOY=true" ||
		got[1] != "REGION=us-east-1" {
		t.Errorf("deploy envs: %#v", got)
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

func TestManifestCobraFields(t *testing.T) {
	yamlDoc := `
command: mycmd
description: Short desc
long_description: |
  Long description line one.
  Line two.
entrypoint: run.sh
use: "<name> [options]"
args: 1
aliases:
  - x
  - run
example: "mb tools mycmd do"
deprecated: "Use 'mb tools newcmd' instead."
`
	var m Manifest
	if err := yaml.Unmarshal([]byte(yamlDoc), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.Command != "mycmd" || m.Description != "Short desc" {
		t.Errorf("command/description: %q / %q", m.Command, m.Description)
	}
	if m.LongDescription != "Long description line one.\nLine two.\n" {
		t.Errorf("long_description: %q", m.LongDescription)
	}
	if m.Use != "<name> [options]" || m.Args != 1 {
		t.Errorf("use=%q args=%d", m.Use, m.Args)
	}
	if len(m.Aliases) != 2 || m.Aliases[0] != "x" || m.Aliases[1] != "run" {
		t.Errorf("aliases: %v", m.Aliases)
	}
	if m.Example != "mb tools mycmd do" {
		t.Errorf("example: %q", m.Example)
	}
	if m.Deprecated != "Use 'mb tools newcmd' instead." {
		t.Errorf("deprecated: %q", m.Deprecated)
	}
}

func TestManifestCobraFieldsEmpty(t *testing.T) {
	yamlDoc := `command: x
description: X
entrypoint: run.sh
`
	var m Manifest
	if err := yaml.Unmarshal([]byte(yamlDoc), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.Use != "" || m.Args != 0 || m.Aliases != nil || m.Example != "" || m.Deprecated != "" ||
		m.LongDescription != "" {
		t.Errorf(
			"cobra fields should be zero: use=%q args=%d aliases=%v example=%q deprecated=%q long_description=%q",
			m.Use,
			m.Args,
			m.Aliases,
			m.Example,
			m.Deprecated,
			m.LongDescription,
		)
	}
}

func TestEnvFilesSpecAndNormalize(t *testing.T) {
	yamlDoc := `
env_files:
  - file: .env
    vault: test
  - file: .env.local
`
	var m Manifest
	if err := yaml.Unmarshal([]byte(yamlDoc), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m.normalizeEnvFileVaults()
	if m.EnvFiles.Len() != 2 {
		t.Fatalf("len %d", m.EnvFiles.Len())
	}
	if m.EnvFiles.List[0].File != ".env" || m.EnvFiles.List[0].Vault != "test" {
		t.Errorf("first: %+v", m.EnvFiles.List[0])
	}
	if m.EnvFiles.List[1].File != ".env.local" ||
		m.EnvFiles.List[1].Vault != ManifestEnvVaultDefault {
		t.Errorf("second: %+v", m.EnvFiles.List[1])
	}
}

func TestEnvFilesSpecMapRejected(t *testing.T) {
	yamlMap := `env_files:
  x: .env
`
	var m Manifest
	err := yaml.Unmarshal([]byte(yamlMap), &m)
	if err == nil {
		t.Fatal("expected error for map env_files")
	}
}

func TestPluginTypeFromEntrypoint(t *testing.T) {
	tests := []struct {
		entrypoint string
		want       string
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
