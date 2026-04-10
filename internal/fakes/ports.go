// Package fakes provides test doubles for the ports used across the application.
// These are designed to be lightweight and composable for unit tests.
package fakes

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"mb/internal/domain/plugin"
	"mb/internal/ports"
)

// --- FakeFilesystem ---

// FakeFS implements ports.Filesystem using an in-memory map.
type FakeFS struct {
	mu      sync.Mutex
	Files   map[string]string // path → content
	Dirs    map[string]bool   // path → true
	Removed []string
}

func NewFakeFS() *FakeFS {
	return &FakeFS{
		Files: make(map[string]string),
		Dirs:  make(map[string]bool),
	}
}

func (f *FakeFS) RemoveAll(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Removed = append(f.Removed, path)
	// Remove files under this path
	for p := range f.Files {
		if strings.HasPrefix(p, path) {
			delete(f.Files, p)
		}
	}
	for p := range f.Dirs {
		if strings.HasPrefix(p, path) {
			delete(f.Dirs, p)
		}
	}
	return nil
}

func (f *FakeFS) MkdirAll(path string, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Dirs[path] = true
	return nil
}

func (f *FakeFS) Stat(name string) (fs.FileInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Dirs[name] {
		return &fakeFileInfo{name: filepath.Base(name), isDir: true}, nil
	}
	if _, ok := f.Files[name]; ok {
		return &fakeFileInfo{name: filepath.Base(name), isDir: false}, nil
	}
	return nil, os.ErrNotExist
}

func (f *FakeFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (f *FakeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries := make(map[string]fs.DirEntry)
	for p := range f.Files {
		dir := filepath.Dir(p)
		if dir == name {
			base := filepath.Base(p)
			entries[base] = &fakeDirEntry{name: base, isDir: false}
		}
	}
	for p := range f.Dirs {
		dir := filepath.Dir(p)
		if dir == name {
			base := filepath.Base(p)
			if base == "." || base == "" {
				continue
			}
			entries[base] = &fakeDirEntry{name: base, isDir: true}
		}
	}
	result := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}

func (f *FakeFS) Getwd() (string, error) {
	return "/tmp", nil
}

// --- fake fs info types ---

type fakeFileInfo struct {
	name  string
	isDir bool
}

func (fi *fakeFileInfo) Name() string               { return fi.name }
func (fi *fakeFileInfo) Size() int64                { return 0 }
func (fi *fakeFileInfo) Mode() fs.FileMode          { return 0o755 }
func (fi *fakeFileInfo) ModTime() time.Time         { return time.Time{} }
func (fi *fakeFileInfo) IsDir() bool                { return fi.isDir }
func (fi *fakeFileInfo) Sys() any                   { return nil }
func (fi *fakeFileInfo) Info() (fs.FileInfo, error) { return fi, nil }
func (fi *fakeFileInfo) Type() fs.FileMode          { return 0 }

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (e *fakeDirEntry) Name() string      { return e.name }
func (e *fakeDirEntry) IsDir() bool       { return e.isDir }
func (e *fakeDirEntry) Type() fs.FileMode { return 0 }
func (e *fakeDirEntry) Info() (fs.FileInfo, error) {
	return &fakeFileInfo{name: e.name, isDir: e.isDir}, nil
}

// --- FakeGitOperations ---

// FakeGit implements ports.GitOperations for testing.
type FakeGit struct {
	ParsedURLs   []string
	ClonedRepos  []string
	ClonedDests  []string
	LatestTags   map[string]string // repoURL → tag
	Versions     map[string]string // dir → version
	Branches     map[string]string // dir → branch
	IsGitRepos   map[string]bool   // dir → isRepo
	CloneErr     error
	LatestTagErr error
}

func NewFakeGit() *FakeGit {
	return &FakeGit{
		LatestTags: make(map[string]string),
		Versions:   make(map[string]string),
		Branches:   make(map[string]string),
		IsGitRepos: make(map[string]bool),
	}
}

func (f *FakeGit) ParseGitURL(raw string) (repoName, normalizedURL string, err error) {
	// Only treat URLs starting with common git schemes as valid Git URLs
	if !strings.HasPrefix(raw, "https://") && !strings.HasPrefix(raw, "git@") &&
		!strings.HasPrefix(raw, "ssh://") && !strings.HasPrefix(raw, "git://") &&
		!strings.HasPrefix(raw, "file://") {
		return "", "", fmt.Errorf("not a git URL: %s", raw)
	}
	f.ParsedURLs = append(f.ParsedURLs, raw)
	parts := strings.Split(raw, "/")
	name := strings.TrimSuffix(parts[len(parts)-1], ".git")
	return name, raw, nil
}

func (f *FakeGit) Clone(
	ctx context.Context,
	repoURL, destDir string,
	opts ports.GitCloneOpts,
) error {
	f.ClonedRepos = append(f.ClonedRepos, repoURL)
	f.ClonedDests = append(f.ClonedDests, destDir)
	return f.CloneErr
}

func (f *FakeGit) LatestTag(ctx context.Context, repoURL string) (string, error) {
	if f.LatestTagErr != nil {
		return "", f.LatestTagErr
	}
	return f.LatestTags[repoURL], nil
}

func (f *FakeGit) GetVersion(dir string) (string, error) {
	if v, ok := f.Versions[dir]; ok {
		return v, nil
	}
	return "1.0.0", nil
}

func (f *FakeGit) GetCurrentBranch(dir string) (string, error) {
	return f.Branches[dir], nil
}

func (f *FakeGit) IsGitRepo(dir string) bool {
	return f.IsGitRepos[dir]
}

func (f *FakeGit) FetchTags(ctx context.Context, dir string) error { return nil }

func (f *FakeGit) ListLocalTags(dir string) ([]string, error) { return nil, nil }

func (f *FakeGit) NewerTag(current, candidate string) (string, bool) { return "", false }

func (f *FakeGit) CheckoutTag(ctx context.Context, dir, tag string) error { return nil }

func (f *FakeGit) FetchAndPull(ctx context.Context, dir, ref string) error { return nil }

// --- FakeShellInstaller ---

// FakeShellInstaller implements ports.ShellHelperInstaller.
type FakeShellInstaller struct {
	CalledWith []string
	EnsureErr  error
}

func (f *FakeShellInstaller) EnsureShellHelpers(configDir string) (string, error) {
	f.CalledWith = append(f.CalledWith, configDir)
	return configDir, f.EnsureErr
}

// --- FakeLayoutValidator ---

// FakeLayoutValidator implements ports.PluginLayoutValidator.
type FakeLayoutValidator struct {
	CalledWith   []string
	ValidateErrs map[string]error
}

func (f *FakeLayoutValidator) ValidatePluginRoot(dir string) error {
	f.CalledWith = append(f.CalledWith, dir)
	return f.ValidateErrs[dir]
}

// --- FakeLogger ---

// FakeLogger captures all log calls for assertions.
type FakeLogger struct {
	mu     sync.Mutex
	Infos  []string
	Warns  []string
	Debugs []string
	Errors []string
}

func NewFakeLogger() *FakeLogger {
	return &FakeLogger{
		Infos:  make([]string, 0),
		Warns:  make([]string, 0),
		Debugs: make([]string, 0),
		Errors: make([]string, 0),
	}
}

func (f *FakeLogger) Info(ctx context.Context, msg string, args ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Infos = append(f.Infos, fmt.Sprintf(msg, args...))
	return nil
}

func (f *FakeLogger) Warn(ctx context.Context, msg string, args ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Warns = append(f.Warns, fmt.Sprintf(msg, args...))
	return nil
}

func (f *FakeLogger) Debug(ctx context.Context, msg string, args ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Debugs = append(f.Debugs, fmt.Sprintf(msg, args...))
	return nil
}

func (f *FakeLogger) Error(ctx context.Context, msg string, args ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Errors = append(f.Errors, fmt.Sprintf(msg, args...))
	return nil
}

// --- FakePluginScanner ---

// FakePluginScanner implements ports.PluginScanner.
type FakePluginScanner struct {
	ScanResult    []plugin.Plugin
	Categories    []plugin.Category
	Warnings      []plugin.ValidationWarning
	HelpGroups    [][]plugin.HelpGroupDef
	ScanErr       error
	DebugLogCalls []string
}

func (f *FakePluginScanner) SetDebugLog(fn func(string)) {
	// Store for tests that need to verify debug log calls
}

func (f *FakePluginScanner) Scan() ([]plugin.Plugin, []plugin.Category, []plugin.ValidationWarning, [][]plugin.HelpGroupDef, error) {
	return f.ScanResult, f.Categories, f.Warnings, f.HelpGroups, f.ScanErr
}

func (f *FakePluginScanner) ScanDir(
	localPath, installDir string,
) ([]plugin.Plugin, []plugin.Category, []plugin.ValidationWarning, [][]plugin.HelpGroupDef, error) {
	return nil, nil, nil, nil, nil
}

// --- Fake helpers ---

var _ ports.Filesystem = (*FakeFS)(nil)
var _ ports.GitOperations = (*FakeGit)(nil)
var _ ports.ShellHelperInstaller = (*FakeShellInstaller)(nil)
var _ ports.PluginLayoutValidator = (*FakeLayoutValidator)(nil)
