package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// CloneOpts configures how to clone a repository.
type CloneOpts struct {
	// BranchOrTag is the branch name or tag to checkout. Empty = default branch (main/master).
	BranchOrTag string
	// UseTag if true, BranchOrTag is a tag (clone in detached HEAD).
	UseTag bool
}

// ParseGitURL parses a Git URL (https or ssh) and returns the repo name (last path segment)
// and a normalized URL for clone. Supports GitHub, Bitbucket, GitLab.
func ParseGitURL(raw string) (repoName string, normalizedURL string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("empty git URL")
	}

	// Normalize: remove .git suffix for comparison
	normalizedURL = raw
	if strings.HasSuffix(normalizedURL, ".git") {
		normalizedURL = strings.TrimSuffix(normalizedURL, ".git")
	}

	var path string
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
		// https://github.com/org/repo or https://github.com/org/repo.git
		parts := strings.SplitN(raw, "//", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid https URL: %s", raw)
		}
		rest := parts[1]
		idx := strings.Index(rest, "/")
		if idx == -1 {
			return "", "", fmt.Errorf("invalid https URL (no path): %s", raw)
		}
		path = rest[idx+1:]
	} else if strings.Contains(raw, "@") && strings.Contains(raw, ":") {
		// git@github.com:org/repo or git@gitlab.com:org/repo.git
		idx := strings.Index(raw, ":")
		if idx == -1 {
			return "", "", fmt.Errorf("invalid ssh URL: %s", raw)
		}
		path = raw[idx+1:]
	} else {
		return "", "", fmt.Errorf("unsupported git URL format: %s", raw)
	}

	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")
	segments := strings.Split(path, "/")
	if len(segments) == 0 || segments[0] == "" {
		return "", "", fmt.Errorf("could not derive repo name from URL: %s", raw)
	}
	repoName = segments[len(segments)-1]
	return repoName, normalizedURL, nil
}

// Clone clones a repository into destDir. If opts.BranchOrTag is set and opts.UseTag is true,
// clones with --branch <tag> (detached). If opts.BranchOrTag is set and UseTag is false,
// clones that branch. If BranchOrTag is empty, clones default branch.
func Clone(ctx context.Context, repoURL string, destDir string, opts CloneOpts) error {
	args := []string{"clone", "--depth", "1"}
	if opts.BranchOrTag != "" {
		args = append(args, "--branch", opts.BranchOrTag)
	}
	args = append(args, repoURL, destDir)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = filepath.Dir(destDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %w: %s", err, string(out))
	}
	return nil
}

// GetCurrentBranch returns the current branch name in dir (e.g. main, master).
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GetVersion returns the current version in dir: tag from git describe, or short SHA.
func GetVersion(dir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		// fallback to short SHA
		cmd2 := exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd2.Dir = dir
		out2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			return "", fmt.Errorf("git describe: %w; rev-parse: %v", err, err2)
		}
		return strings.TrimSpace(string(out2)), nil
	}
	return strings.TrimSpace(string(out)), nil
}

// IsGitRepo returns true if dir is a git repository.
func IsGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

// LatestTag returns the latest tag from the remote (by semver order). Uses ls-remote.
func LatestTag(ctx context.Context, repoURL string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", repoURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git ls-remote --tags: %w: %s", err, string(out))
	}

	tags := parseTagsFromLsRemote(string(out))
	if len(tags) == 0 {
		return "", nil
	}

	sort.Slice(tags, func(i, j int) bool {
		return compareSemver(tags[i], tags[j]) > 0
	})
	return tags[0], nil
}

func parseTagsFromLsRemote(out string) []string {
	var tags []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// format: SHA\trefs/tags/v1.2.3 or refs/tags/v1.2.3^{}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := fields[1]
		if !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}
		tag := strings.TrimPrefix(ref, "refs/tags/")
		if strings.HasSuffix(tag, "^{}") {
			tag = strings.TrimSuffix(tag, "^{}")
		}
		tags = append(tags, tag)
	}
	return tags
}

// compareSemver returns >0 if a is greater than b, 0 if equal, <0 if a < b.
// Non-semver tags are compared lexicographically after semver ones.
func compareSemver(a, b string) int {
	va, na := parseSemver(a)
	vb, nb := parseSemver(b)
	if na && nb {
		if va[0] != vb[0] {
			return va[0] - vb[0]
		}
		if len(va) > 1 && len(vb) > 1 {
			if va[1] != vb[1] {
				return va[1] - vb[1]
			}
		}
		if len(va) > 2 && len(vb) > 2 {
			return va[2] - vb[2]
		}
		return len(va) - len(vb)
	}
	if na {
		return 1
	}
	if nb {
		return -1
	}
	return strings.Compare(a, b)
}

var semverRe = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// parseSemver extracts major, minor, patch from a tag. Returns (nil, false) if not semver.
func parseSemver(tag string) ([]int, bool) {
	m := semverRe.FindStringSubmatch(tag)
	if m == nil {
		return nil, false
	}
	var v []int
	for i := 1; i < len(m); i++ {
		if m[i] == "" {
			v = append(v, 0)
			continue
		}
		var n int
		for _, c := range m[i] {
			n = n*10 + int(c-'0')
		}
		v = append(v, n)
	}
	return v, true
}

// FetchTags runs git fetch --tags in dir.
func FetchTags(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "fetch", "--tags")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch --tags: %w: %s", err, string(out))
	}
	return nil
}

// ListLocalTags returns tags in dir, sorted by semver descending.
func ListLocalTags(dir string) ([]string, error) {
	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git tag -l: %w: %s", err, string(out))
	}
	tags := strings.Fields(string(out))
	sort.Slice(tags, func(i, j int) bool {
		return compareSemver(tags[i], tags[j]) > 0
	})
	return tags, nil
}

// CheckoutTag checks out the given tag in dir.
func CheckoutTag(ctx context.Context, dir string, tag string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", "tags/"+tag)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout tags/%s: %w: %s", tag, err, string(out))
	}
	return nil
}

// FetchAndPull runs git fetch and git pull for the given ref (branch) in dir.
func FetchAndPull(ctx context.Context, dir string, ref string) error {
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin", ref)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch origin %s: %w: %s", ref, err, string(out))
	}
	cmd2 := exec.CommandContext(ctx, "git", "checkout", ref)
	cmd2.Dir = dir
	if out, err := cmd2.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout %s: %w: %s", ref, err, string(out))
	}
	cmd3 := exec.CommandContext(ctx, "git", "pull", "origin", ref)
	cmd3.Dir = dir
	if out, err := cmd3.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull origin %s: %w: %s", ref, err, string(out))
	}
	return nil
}

// NewerTag returns the newer of two tags (by semver). If a is newer or equal, returns a; else returns b.
func NewerTag(current, candidate string) (newer string, isNewer bool) {
	cmp := compareSemver(candidate, current)
	if cmp > 0 {
		return candidate, true
	}
	return current, false
}
