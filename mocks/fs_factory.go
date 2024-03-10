package mocks

import (
	"os"
	"path"
	"time"
)

type statFunc func(filename string) (os.FileInfo, error)

// FsWrapper wraps numerous file system interactions
type FsWrapper struct {
	Stat    statFunc
	ReadDir func(dir string) ([]os.DirEntry, error)
	Open    func(filename string) (*os.File, error)
}

var fsWrapper *FsWrapper

type MockFileInfo struct {
	name    string
	size    int64
	isDir   bool
	mode    os.FileMode
	modTime time.Time
}

func (i MockFileInfo) Name() string {
	return i.name
}
func (i MockFileInfo) Size() int64 {
	return i.size
}
func (i MockFileInfo) Mode() os.FileMode {
	return i.mode
}
func (i MockFileInfo) ModTime() time.Time {
	return i.modTime
}
func (i MockFileInfo) IsDir() bool {
	return i.isDir
}
func (i MockFileInfo) Sys() any {
	return nil
}

type MockDirEntry struct {
	IName  string
	IPath  string
	IIsDir bool
	IMode  os.FileMode
	IStat  statFunc
}

func (e MockDirEntry) Name() string {
	return e.IName
}
func (e MockDirEntry) Type() os.FileMode {
	return e.IMode
}
func (e MockDirEntry) IsDir() bool {
	return e.IIsDir
}
func (e MockDirEntry) Info() (os.FileInfo, error) {
	return e.IStat(path.Join(e.IPath, e.IName))
}

// GetFsWrapper returns default implementations for FS functions
func GetFsWrapper() *FsWrapper {
	if fsWrapper == nil {
		fsWrapper = &FsWrapper{
			Stat:    os.Stat,
			ReadDir: os.ReadDir,
			Open:    os.Open,
		}
	}
	return fsWrapper
}
