package plugincmd

import (
	"testing"

	"mb/internal/deps"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

func TestNewLeafCommand_FlagsWithReadmeNoPanicWhenPluginUsesR(t *testing.T) {
	flagsJSON := `{"run":{"type":"long","short":"r","entrypoint":"run.sh","description":"run"}}`
	plugin := sqlite.Plugin{
		CommandPath: "dev/bump",
		CommandName: "bump",
		FlagsJSON:   flagsJSON,
		ReadmePath:  "/tmp/mb-readme-test.md",
	}
	d := deps.NewDependencies(&deps.RuntimeConfig{}, config.AppConfig{}, nil, nil, nil)
	cmd := newLeafCommand("bump", plugin, d, "/tmp", false, nil, nil)
	rf := cmd.Flags().Lookup("readme")
	if rf == nil {
		t.Fatal("readme flag missing")
	}
	if rf.Shorthand != "r" {
		t.Fatalf(
			"readme should keep -r (MB registado antes dos flags do plugin), got shorthand %q",
			rf.Shorthand,
		)
	}
	runF := cmd.Flags().Lookup("run")
	if runF == nil || runF.Shorthand != "" {
		t.Fatalf("run flag should be long-only when -r fica com --readme, got %#v", runF)
	}
}

func TestNewLeafCommand_ReservedRootShorthandDropped(t *testing.T) {
	flagsJSON := `{"watch":{"type":"long","short":"v","entrypoint":"w.sh","description":"w"}}`
	plugin := sqlite.Plugin{
		CommandPath: "p/x",
		CommandName: "x",
		FlagsJSON:   flagsJSON,
	}
	d := deps.NewDependencies(&deps.RuntimeConfig{}, config.AppConfig{}, nil, nil, nil)
	global := map[string]struct{}{"v": {}}
	cmd := newLeafCommand("x", plugin, d, "/tmp", false, nil, global)
	f := cmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("watch flag missing")
	}
	if f.Shorthand != "" {
		t.Fatalf("want empty shorthand when manifest requests reserved v, got %q", f.Shorthand)
	}
}
