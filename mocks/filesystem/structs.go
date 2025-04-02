package filesystem

import (
	"os"
	"path"
	"time"
)

type statFunc func(filename string) (os.FileInfo, error)

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
	FileName string
	FilePath string
	Dir      bool
	FileMode os.FileMode
	StatFunc statFunc
}

func (e MockDirEntry) Name() string {
	return e.FileName
}
func (e MockDirEntry) Type() os.FileMode {
	return e.FileMode
}
func (e MockDirEntry) IsDir() bool {
	return e.Dir
}
func (e MockDirEntry) Info() (os.FileInfo, error) {
	return e.StatFunc(path.Join(e.FilePath, e.FileName))
}
