package plugin

import "testing"

func TestMergeHelpGroupsGlobal_conflict(t *testing.T) {
	var msgs []string
	out := MergeHelpGroupsGlobal([][]HelpGroupDef{
		{{ID: "a", Title: "First"}},
		{{ID: "a", Title: "Second"}},
	}, func(m string) { msgs = append(msgs, m) })
	if len(out) != 1 || out[0].Title != "Second" {
		t.Fatalf("out %+v", out)
	}
	if len(msgs) != 1 {
		t.Fatalf("msgs %v", msgs)
	}
}
