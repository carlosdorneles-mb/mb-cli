package fs

import (
	"io/fs"
	"os"
)

// OS implements ports.Filesystem using the real operating system.
type OS struct{}

func (OS) RemoveAll(path string) error { return os.RemoveAll(path) }

func (OS) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

func (OS) IsNotExist(err error) bool { return os.IsNotExist(err) }

func (OS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (OS) Getwd() (string, error) { return os.Getwd() }
