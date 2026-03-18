package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanTreeInfraStyleCommandPaths(t *testing.T) {
	root := t.TempDir()
	// Root category manifest (like examples/plugins/infra)
	if err := os.WriteFile(
		filepath.Join(root, "manifest.yaml"),
		[]byte("command: infra\ndescription: root\nreadme: README.md\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// ci/category
	if err := os.MkdirAll(filepath.Join(root, "ci", "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "ci", "manifest.yaml"),
		[]byte("command: ci\ndescription: CI\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "ci", "deploy", "manifest.yaml"),
		[]byte("command: deploy\ndescription: d\nentrypoint: run.sh\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "ci", "deploy", "run.sh"),
		[]byte("#!/bin/sh\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	s := NewScanner(filepath.Join(root, "unused"))
	p, cats, w, _, err := s.scanTree(root)
	if err != nil {
		t.Fatalf("scanTree: %v", err)
	}
	if len(w) != 0 {
		t.Errorf("warnings: %+v", w)
	}
	var paths []string
	for _, pl := range p {
		paths = append(paths, pl.CommandPath)
	}
	wantPath := "infra/ci/deploy"
	found := false
	for _, pth := range paths {
		if pth == wantPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("want CommandPath %q among %#v", wantPath, paths)
	}
	var catPaths []string
	for _, c := range cats {
		catPaths = append(catPaths, c.Path)
	}
	for _, need := range []string{"infra", "infra/ci"} {
		if !containsString(catPaths, need) {
			t.Errorf("missing category %q in %#v", need, catPaths)
		}
	}
}

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func TestScanTreeSinglePluginNoPrefix(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "manifest.yaml"),
		[]byte("command: hello\ndescription: x\nentrypoint: run.sh\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "run.sh"),
		[]byte("#!/bin/sh\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	s := NewScanner("/tmp")
	p, _, _, _, err := s.scanTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 {
		t.Fatalf("plugins: %+v", p)
	}
	// rel "." leaves commandPath empty; entrypoint fills dbCommandPath with command name
	if p[0].CommandPath != "hello" || p[0].CommandName != "hello" {
		t.Errorf("leaf at root: CommandPath=%q CommandName=%q", p[0].CommandPath, p[0].CommandName)
	}
}

func TestScanTreeNestedGroupID(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "manifest.yaml"),
		[]byte("command: infra\ndescription: root\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "svc", "run"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "svc", "groups.yaml"), []byte(`
- id: my_grp
  title: MY SECTION
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "svc", "manifest.yaml"),
		[]byte("command: svc\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "svc", "run", "manifest.yaml"),
		[]byte("command: run\ndescription: r\ngroup_id: my_grp\nentrypoint: x.sh\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "svc", "run", "x.sh"),
		[]byte("#!/bin/sh\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	s := NewScanner("/tmp")
	p, _, _, _, err := s.scanTree(root)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	for _, pl := range p {
		if pl.CommandPath == "infra/svc/run" {
			got = pl.GroupID
			break
		}
	}
	if got != "my_grp" {
		t.Errorf("GroupID=%q want my_grp", got)
	}
}

func TestScanTreeNestedPreservesGroupIDFromAnyDepth(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(
		filepath.Join(root, "manifest.yaml"),
		[]byte("command: pkg\ndescription: p\n"),
		0o644,
	)
	_ = os.WriteFile(
		filepath.Join(root, "groups.yaml"),
		[]byte("- id: root_only\n  title: R\n"),
		0o644,
	)
	_ = os.MkdirAll(filepath.Join(root, "a", "b"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "a", "manifest.yaml"), []byte("command: a\n"), 0o644)
	_ = os.WriteFile(
		filepath.Join(root, "a", "b", "manifest.yaml"),
		[]byte("command: leaf\ngroup_id: root_only\nentrypoint: x.sh\n"),
		0o644,
	)
	_ = os.WriteFile(filepath.Join(root, "a", "b", "x.sh"), []byte("#!/bin/sh\n"), 0o755)
	s := NewScanner("/tmp")
	p, _, _, _, err := s.scanTree(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, pl := range p {
		if pl.CommandName == "leaf" {
			if pl.GroupID != "root_only" {
				t.Errorf("scanner keeps raw group_id; want root_only, got %q", pl.GroupID)
			}
		}
	}
}

func TestScanTreeTopLevelIgnoresGroupID(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "groups.yaml"),
		[]byte("- id: g\n  title: G\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "manifest.yaml"),
		[]byte("command: hello\ndescription: h\ngroup_id: g\nentrypoint: run.sh\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "run.sh"),
		[]byte("#!/bin/sh\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	var debug []string
	s := NewScanner("/tmp")
	s.DebugLog = func(msg string) { debug = append(debug, msg) }
	p, _, _, _, err := s.scanTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 {
		t.Fatalf("plugins %+v", p)
	}
	if p[0].CommandPath != "hello" || p[0].GroupID != "" {
		t.Errorf("leaf: CommandPath=%q GroupID=%q", p[0].CommandPath, p[0].GroupID)
	}
	if len(debug) != 1 {
		t.Errorf("want 1 debug line, got %d: %v", len(debug), debug)
	}
}

func TestValidatePluginRoot(t *testing.T) {
	root := t.TempDir()
	err := ValidatePluginRoot(root)
	if err == nil || !strings.Contains(err.Error(), "manifest") {
		t.Errorf("expected error without manifest: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "manifest.yaml"),
		[]byte("command: x\ndescription: y\nentrypoint: run.sh\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePluginRoot(root); err == nil || !strings.Contains(err.Error(), "entrypoint") {
		t.Errorf("missing run.sh: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "run.sh"),
		[]byte("#!/bin/sh\n"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePluginRoot(root); err != nil {
		t.Error(err)
	}
}
