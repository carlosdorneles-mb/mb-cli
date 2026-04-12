package addplugin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mb/internal/domain/plugin"
	"mb/internal/fakes"
	"mb/internal/ports"
)

// --- Tests for isAuthError ---

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		wantAuth bool
	}{
		{"HTTP 401", "git clone: exit status 128: 401 Unauthorized", true},
		{"Authentication failed", "git clone: exit status 128: Authentication failed", true},
		{"Could not read", "git clone: exit status 128: could not read Username", true},
		{"Permission denied", "git clone: exit status 1: Permission denied (publickey)", true},
		{"Invalid username", "remote: Invalid username or password", true},
		{"HTTP Basic denied", "HTTP Basic: Access denied", true},
		{"Fatal could not read", "fatal: could not read Password", true},
		{"Remote auth required", "remote: Authentication required", true},
		{"Remote unauthorized", "remote: Unauthorized", true},
		{"Access denied", "remote: Access denied", true},
		{"Repo not found", "repository 'xyz' not found", false},
		{"Network error", "git clone: exit status 128: Connection refused", false},
		{"Nil error", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = errors.New(tt.errMsg)
			}
			got := isAuthError(err)
			if got != tt.wantAuth {
				t.Errorf("isAuthError(%q) = %v, want %v", tt.errMsg, got, tt.wantAuth)
			}
		})
	}
}

// --- Tests for generateFallbackURLs ---

func TestGenerateFallbackURLs(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name          string
		normalizedURL string
		originalURL   string
		wantURLs      []string
	}{
		{
			name:          "HTTPS without .git adds .git and SSH",
			normalizedURL: "https://github.com/org/repo",
			originalURL:   "https://github.com/org/repo",
			wantURLs: []string{
				"https://github.com/org/repo",
				"https://github.com/org/repo.git",
				"git@github.com:org/repo.git",
			},
		},
		{
			name:          "HTTPS with .git adds only SSH",
			normalizedURL: "https://github.com/org/repo.git",
			originalURL:   "https://github.com/org/repo.git",
			wantURLs: []string{
				"https://github.com/org/repo.git",
				"git@github.com:org/repo.git",
			},
		},
		{
			name:          "HTTP without .git",
			normalizedURL: "http://gitlab.com/org/repo",
			originalURL:   "http://gitlab.com/org/repo",
			wantURLs: []string{
				"http://gitlab.com/org/repo",
				"http://gitlab.com/org/repo.git",
				"git@gitlab.com:org/repo.git",
			},
		},
		{
			name:          "SSH URL stays as-is",
			normalizedURL: "git@github.com:org/repo",
			originalURL:   "git@github.com:org/repo",
			wantURLs: []string{
				"git@github.com:org/repo",
				"git@github.com:org/repo.git",
			},
		},
		{
			name:          "SSH URL with .git stays as-is",
			normalizedURL: "git@github.com:org/repo.git",
			originalURL:   "git@github.com:org/repo.git",
			wantURLs: []string{
				"git@github.com:org/repo.git",
			},
		},
		{
			name:          "Subgroup GitLab HTTPS",
			normalizedURL: "https://gitlab.com/org/subgroup/repo",
			originalURL:   "https://gitlab.com/org/subgroup/repo",
			wantURLs: []string{
				"https://gitlab.com/org/subgroup/repo",
				"https://gitlab.com/org/subgroup/repo.git",
				"git@gitlab.com:org/subgroup/repo.git",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.generateFallbackURLs(tt.normalizedURL)
			if len(got) != len(tt.wantURLs) {
				t.Fatalf("got %d URLs, want %d: %v", len(got), len(tt.wantURLs), got)
			}
			for i, u := range got {
				if u != tt.wantURLs[i] {
					t.Errorf("[%d] = %q, want %q", i, u, tt.wantURLs[i])
				}
			}
		})
	}
}

// --- Integration-style test for addRemote with fallback ---

func TestAddRemote_FallbackSuccessOnSecondAttempt(t *testing.T) {
	fsys := fakes.NewFakeFS()
	shell := &fakes.FakeShellInstaller{}
	layout := &fakes.FakeLayoutValidator{}
	logger := fakes.NewFakeLogger()

	store := newFakeStoreForAdd()
	scanner := &fakes.FakePluginScanner{}
	syncer := NewSyncer()

	// FakeGit that fails with auth error on HTTPS URLs, succeeds on SSH
	git := &fakeGitWithFallback{
		failOnURLs: map[string]bool{
			"https://github.com/org/repo":     true,
			"https://github.com/org/repo.git": true,
		},
		cloneErr: errors.New("git clone: exit status 128: 401 Unauthorized"),
	}

	rt := Runtime{
		ConfigDir:  "/tmp/config",
		PluginsDir: "/tmp/plugins",
	}
	_ = fsys.MkdirAll(rt.PluginsDir, 0o755)

	svc := New(rt, store, scanner, fsys, git, shell, layout, syncer)

	err := svc.addRemote(
		t.Context(),
		"https://github.com/org/repo",
		"repo",
		"",
		SyncOptions{EmitSuccess: false},
		logger,
	)

	if err != nil {
		t.Fatalf("addRemote should succeed with fallback: %v", err)
	}

	// Verify it tried the SSH fallback URL
	if !git.triedSSHURL {
		t.Error("expected service to try SSH fallback")
	}
}

func TestAddRemote_AllAttemptsFailWithAuthError(t *testing.T) {
	fsys := fakes.NewFakeFS()
	shell := &fakes.FakeShellInstaller{}
	layout := &fakes.FakeLayoutValidator{}
	logger := fakes.NewFakeLogger()

	store := newFakeStoreForAdd()
	scanner := &fakes.FakePluginScanner{}
	syncer := NewSyncer()

	// FakeGit that always fails with auth error
	git := &fakeGitAlwaysFails{
		err: errors.New("git clone: exit status 128: 401 Unauthorized"),
	}

	rt := Runtime{
		ConfigDir:  "/tmp/config",
		PluginsDir: "/tmp/plugins",
	}
	_ = fsys.MkdirAll(rt.PluginsDir, 0o755)

	svc := New(rt, store, scanner, fsys, git, shell, layout, syncer)

	err := svc.addRemote(
		t.Context(),
		"https://github.com/org/repo",
		"repo",
		"",
		SyncOptions{EmitSuccess: false},
		logger,
	)

	if err == nil {
		t.Fatal("expected error when all attempts fail")
	}
	if !strings.Contains(err.Error(), "não tem permissão") {
		t.Errorf("expected permission error, got: %v", err)
	}
}

func TestAddRemote_NonAuthErrorFailsImmediately(t *testing.T) {
	fsys := fakes.NewFakeFS()
	shell := &fakes.FakeShellInstaller{}
	layout := &fakes.FakeLayoutValidator{}
	logger := fakes.NewFakeLogger()

	store := newFakeStoreForAdd()
	scanner := &fakes.FakePluginScanner{}
	syncer := NewSyncer()

	// FakeGit that fails with non-auth error
	git := &fakeGitAlwaysFails{
		err: errors.New("git clone: exit status 128: repository not found"),
	}

	rt := Runtime{
		ConfigDir:  "/tmp/config",
		PluginsDir: "/tmp/plugins",
	}
	_ = fsys.MkdirAll(rt.PluginsDir, 0o755)

	svc := New(rt, store, scanner, fsys, git, shell, layout, syncer)

	err := svc.addRemote(
		t.Context(),
		"https://github.com/org/repo",
		"repo",
		"",
		SyncOptions{EmitSuccess: false},
		logger,
	)

	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "não tem permissão") {
		t.Errorf("should NOT show permission error for non-auth failure: %v", err)
	}
	if !strings.Contains(err.Error(), "clone:") {
		t.Errorf("expected clone error, got: %v", err)
	}
}

// --- Fake implementations ---

// fakeGitWithFallback fails on specific URLs, succeeds on SSH URLs.
type fakeGitWithFallback struct {
	failOnURLs  map[string]bool
	cloneErr    error
	triedSSHURL bool
}

func (f *fakeGitWithFallback) ParseGitURL(raw string) (string, string, error) {
	parts := strings.Split(raw, "/")
	name := strings.TrimSuffix(parts[len(parts)-1], ".git")
	return name, raw, nil
}

func (f *fakeGitWithFallback) Clone(
	_ context.Context, repoURL, _ string, _ ports.GitCloneOpts,
) error {
	if f.failOnURLs[repoURL] {
		return f.cloneErr
	}
	if strings.HasPrefix(repoURL, "git@") {
		f.triedSSHURL = true
	}
	return nil
}

func (f *fakeGitWithFallback) LatestTag(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (f *fakeGitWithFallback) GetVersion(string) (string, error) { return "1.0.0", nil }

func (f *fakeGitWithFallback) GetCurrentBranch(string) (string, error) { return "main", nil }

func (f *fakeGitWithFallback) IsGitRepo(string) bool { return true }

func (f *fakeGitWithFallback) FetchTags(context.Context, string) error { return nil }

func (f *fakeGitWithFallback) ListLocalTags(string) ([]string, error) { return nil, nil }

func (f *fakeGitWithFallback) NewerTag(string, string) (string, bool) { return "", false }

func (f *fakeGitWithFallback) CheckoutTag(context.Context, string, string) error { return nil }

func (f *fakeGitWithFallback) FetchAndPull(context.Context, string, string) error { return nil }

// fakeGitAlwaysFails always returns error on Clone.
type fakeGitAlwaysFails struct {
	err error
}

func (f *fakeGitAlwaysFails) ParseGitURL(raw string) (string, string, error) {
	parts := strings.Split(raw, "/")
	name := strings.TrimSuffix(parts[len(parts)-1], ".git")
	return name, raw, nil
}

func (f *fakeGitAlwaysFails) Clone(context.Context, string, string, ports.GitCloneOpts) error {
	return f.err
}

func (f *fakeGitAlwaysFails) LatestTag(context.Context, string) (string, error) { return "", nil }

func (f *fakeGitAlwaysFails) GetVersion(string) (string, error) { return "1.0.0", nil }

func (f *fakeGitAlwaysFails) GetCurrentBranch(string) (string, error) { return "main", nil }

func (f *fakeGitAlwaysFails) IsGitRepo(string) bool { return true }

func (f *fakeGitAlwaysFails) FetchTags(context.Context, string) error { return nil }

func (f *fakeGitAlwaysFails) ListLocalTags(string) ([]string, error) { return nil, nil }

func (f *fakeGitAlwaysFails) NewerTag(string, string) (string, bool) { return "", false }

func (f *fakeGitAlwaysFails) CheckoutTag(context.Context, string, string) error { return nil }

func (f *fakeGitAlwaysFails) FetchAndPull(context.Context, string, string) error { return nil }

// Ensure fakes satisfy the interface
var _ ports.GitOperations = (*fakeGitWithFallback)(nil)
var _ ports.GitOperations = (*fakeGitAlwaysFails)(nil)

// fakeStoreForAdd implements ports.PluginCacheStore for unit tests.
type fakeStoreForAdd struct {
	sources    map[string]plugin.PluginSource
	plugins    []plugin.Plugin
	categories []plugin.Category
	helpGroups []plugin.PluginHelpGroup
}

func newFakeStoreForAdd() *fakeStoreForAdd {
	return &fakeStoreForAdd{
		sources: make(map[string]plugin.PluginSource),
	}
}

func (f *fakeStoreForAdd) GetPluginSource(id string) (*plugin.PluginSource, error) {
	if s, ok := f.sources[id]; ok {
		return &s, nil
	}
	return nil, nil
}

func (f *fakeStoreForAdd) UpsertPluginSource(s plugin.PluginSource) error {
	src := s
	f.sources[s.InstallDir] = src
	return nil
}

func (f *fakeStoreForAdd) ListPluginSources() ([]plugin.PluginSource, error) {
	result := make([]plugin.PluginSource, 0, len(f.sources))
	for _, s := range f.sources {
		result = append(result, s)
	}
	return result, nil
}

func (f *fakeStoreForAdd) DeletePluginSource(id string) error {
	delete(f.sources, id)
	return nil
}

func (f *fakeStoreForAdd) DeleteAllPluginSources() error {
	f.sources = make(map[string]plugin.PluginSource)
	return nil
}

func (f *fakeStoreForAdd) ListPlugins() ([]plugin.Plugin, error) { return f.plugins, nil }
func (f *fakeStoreForAdd) DeleteAllPlugins() error               { f.plugins = nil; return nil }
func (f *fakeStoreForAdd) DeleteAllPluginHelpGroups() error      { f.helpGroups = nil; return nil }
func (f *fakeStoreForAdd) UpsertPluginHelpGroup(g plugin.PluginHelpGroup) error {
	f.helpGroups = append(f.helpGroups, g)
	return nil
}
func (f *fakeStoreForAdd) UpsertPlugin(p plugin.Plugin) error {
	f.plugins = append(f.plugins, p)
	return nil
}
func (f *fakeStoreForAdd) DeleteAllCategories() error { f.categories = nil; return nil }
func (f *fakeStoreForAdd) UpsertCategory(c plugin.Category) error {
	f.categories = append(f.categories, c)
	return nil
}
func (f *fakeStoreForAdd) ListCategories() ([]plugin.Category, error) { return f.categories, nil }
func (f *fakeStoreForAdd) ListPluginHelpGroups() ([]plugin.PluginHelpGroup, error) {
	return f.helpGroups, nil
}
