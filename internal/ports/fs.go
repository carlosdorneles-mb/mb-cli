package ports

import "io/fs"

// Filesystem is a narrow OS surface for plugin add/remove (testable fakes).
type Filesystem interface {
	RemoveAll(path string) error
	MkdirAll(path string, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	IsNotExist(err error) bool
	ReadDir(name string) ([]fs.DirEntry, error)
	Getwd() (string, error)
}
