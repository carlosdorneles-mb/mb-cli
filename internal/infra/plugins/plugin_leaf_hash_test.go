package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginLeafConfigHash_stableWhenUnchanged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	raw := []byte("command: x\nentrypoint: run.sh\n")
	if err := os.WriteFile(manifestPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	runPath := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(runPath, []byte("#!/bin/sh\necho hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	h1, err := PluginLeafConfigHash(raw, &m, dir)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := PluginLeafConfigHash(raw, &m, dir)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("same inputs: got %s and %s", h1, h2)
	}
}

func TestPluginLeafConfigHash_changesWhenReferencedScriptChanges(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	raw := []byte("command: x\nentrypoint: run.sh\n")
	if err := os.WriteFile(manifestPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	runPath := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(runPath, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	before, err := PluginLeafConfigHash(raw, &m, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runPath, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := PluginLeafConfigHash(raw, &m, dir)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("hash should change when only run.sh changes")
	}
}

func TestPluginLeafConfigHash_changesWhenManifestChanges(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runPath := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(runPath, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw1 := []byte("command: x\nentrypoint: run.sh\n")
	manifestPath := filepath.Join(dir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, raw1, 0o644); err != nil {
		t.Fatal(err)
	}
	m1, _, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	h1, err := PluginLeafConfigHash(raw1, &m1, dir)
	if err != nil {
		t.Fatal(err)
	}
	raw2 := []byte("command: x\nentrypoint: run.sh\n# comment\n")
	if err := os.WriteFile(manifestPath, raw2, 0o644); err != nil {
		t.Fatal(err)
	}
	m2, _, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := PluginLeafConfigHash(raw2, &m2, dir)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Fatal("hash should change when manifest raw changes")
	}
}
