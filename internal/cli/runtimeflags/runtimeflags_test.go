package runtimeflags

import (
	"testing"

	"mb/internal/deps"
)

func TestParseLeadingRuntimeFlags_NoFlags(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	rest, err := ParseLeadingRuntimeFlags(rt, []string{"grep", "-r", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 3 || rest[0] != "grep" {
		t.Fatalf("rest=%v", rest)
	}
}

func TestParseLeadingRuntimeFlags_PrefixEnvAndVerbose(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	rest, err := ParseLeadingRuntimeFlags(rt, []string{"-e", "K=v", "--verbose", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 2 || rest[0] != "echo" {
		t.Fatalf("rest=%v", rest)
	}
	if len(rt.InlineEnvValues) != 1 || rt.InlineEnvValues[0] != "K=v" {
		t.Fatalf("inline=%v", rt.InlineEnvValues)
	}
	if !rt.Verbose {
		t.Fatal("verbose not set")
	}
}

func TestParseLeadingRuntimeFlags_EnvEqualsForm(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	rest, err := ParseLeadingRuntimeFlags(rt, []string{"--env=PATH=/a=b", "true"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 1 || rest[0] != "true" {
		t.Fatalf("rest=%v", rest)
	}
	if len(rt.InlineEnvValues) != 1 || rt.InlineEnvValues[0] != "PATH=/a=b" {
		t.Fatalf("got %v", rt.InlineEnvValues)
	}
}

func TestParseLeadingRuntimeFlags_DoubleDash(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	rest, err := ParseLeadingRuntimeFlags(rt, []string{"--", "--env-vault", "x", "sh"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 3 || rest[0] != "--env-vault" {
		t.Fatalf("rest=%v", rest)
	}
}

func TestParseLeadingRuntimeFlags_MergesWithExisting(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{Verbose: true, InlineEnvValues: []string{"A=1"}}
	rest, err := ParseLeadingRuntimeFlags(rt, []string{"-e", "B=2", "true"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 1 {
		t.Fatalf("rest=%v", rest)
	}
	if !rt.Verbose {
		t.Fatal("verbose lost")
	}
	if len(rt.InlineEnvValues) != 2 || rt.InlineEnvValues[0] != "A=1" ||
		rt.InlineEnvValues[1] != "B=2" {
		t.Fatalf("inline=%v", rt.InlineEnvValues)
	}
}

func TestParseLeadingRuntimeFlags_EnvFileGroup(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	rest, err := ParseLeadingRuntimeFlags(
		rt,
		[]string{"--env-file=/tmp/x.env", "--env-vault=staging", "cmd"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 1 || rest[0] != "cmd" {
		t.Fatalf("rest=%v", rest)
	}
	if rt.EnvFilePath != "/tmp/x.env" || rt.EnvVault != "staging" {
		t.Fatalf("env-file=%q vault=%q", rt.EnvFilePath, rt.EnvVault)
	}
}

func TestParseLeadingRuntimeFlags_UnknownFlag(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	_, err := ParseLeadingRuntimeFlags(rt, []string{"--not-a-real-flag", "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseLeadingRuntimeFlags_EMissingValue(t *testing.T) {
	t.Parallel()
	rt := &deps.RuntimeConfig{}
	_, err := ParseLeadingRuntimeFlags(rt, []string{"-e"})
	if err == nil {
		t.Fatal("expected error")
	}
}
