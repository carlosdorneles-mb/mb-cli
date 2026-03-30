package ports

import "context"

// GitCloneOpts configures a shallow clone (branch or tag).
type GitCloneOpts struct {
	BranchOrTag string
	UseTag      bool
}

// GitOperations abstracts git commands used to install and update remote plugins.
type GitOperations interface {
	ParseGitURL(raw string) (repoName, normalizedURL string, err error)
	Clone(ctx context.Context, repoURL, destDir string, opts GitCloneOpts) error
	LatestTag(ctx context.Context, repoURL string) (string, error)
	GetVersion(dir string) (string, error)
	GetCurrentBranch(dir string) (string, error)
	IsGitRepo(dir string) bool
	FetchTags(ctx context.Context, dir string) error
	ListLocalTags(dir string) ([]string, error)
	NewerTag(current, candidate string) (newer string, isNewer bool)
	CheckoutTag(ctx context.Context, dir, tag string) error
	FetchAndPull(ctx context.Context, dir, ref string) error
}
