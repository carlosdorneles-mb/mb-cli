package safepath

import (
	"errors"
	"path/filepath"
	"strings"
)

// ErrPathOutsideDir is returned when a path is not under the allowed directory.
var ErrPathOutsideDir = errors.New("path is outside allowed directory")

// PathUnderDir reports whether path is under dir (or equal to dir).
// Both path and dir are resolved to absolute, cleaned paths before comparison.
// Path traversal (e.g. "..") in path is evaluated; if the result leaves dir, it returns false.
func PathUnderDir(path, dir string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	absPath = filepath.Clean(absPath)
	absDir = filepath.Clean(absDir)
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false, err
	}
	// Rel can produce ".." or ".." prefix if path is outside dir. On Windows it might use \.
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}

// ValidateUnderDir returns nil if path is under dir, otherwise ErrPathOutsideDir.
func ValidateUnderDir(path, dir string) error {
	ok, err := PathUnderDir(path, dir)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPathOutsideDir
	}
	return nil
}
