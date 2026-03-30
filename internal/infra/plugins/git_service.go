package plugins

import (
	"context"

	"mb/internal/ports"
)

// GitService implements ports.GitOperations using package-level git helpers.
type GitService struct{}

func (GitService) ParseGitURL(raw string) (repoName, normalizedURL string, err error) {
	return ParseGitURL(raw)
}

func (GitService) Clone(
	ctx context.Context,
	repoURL, destDir string,
	opts ports.GitCloneOpts,
) error {
	return Clone(ctx, repoURL, destDir, CloneOpts{
		BranchOrTag: opts.BranchOrTag,
		UseTag:      opts.UseTag,
	})
}

func (GitService) LatestTag(ctx context.Context, repoURL string) (string, error) {
	return LatestTag(ctx, repoURL)
}

func (GitService) GetVersion(dir string) (string, error) { return GetVersion(dir) }

func (GitService) GetCurrentBranch(dir string) (string, error) { return GetCurrentBranch(dir) }

func (GitService) IsGitRepo(dir string) bool { return IsGitRepo(dir) }

func (GitService) FetchTags(ctx context.Context, dir string) error { return FetchTags(ctx, dir) }

func (GitService) ListLocalTags(dir string) ([]string, error) { return ListLocalTags(dir) }

func (GitService) NewerTag(current, candidate string) (string, bool) {
	return NewerTag(current, candidate)
}

func (GitService) CheckoutTag(ctx context.Context, dir, tag string) error {
	return CheckoutTag(ctx, dir, tag)
}

func (GitService) FetchAndPull(ctx context.Context, dir, ref string) error {
	return FetchAndPull(ctx, dir, ref)
}
