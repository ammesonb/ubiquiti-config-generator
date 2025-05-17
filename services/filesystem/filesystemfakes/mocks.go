package filesystemfakes

import (
	"os"
	"path"
	"time"
)

// ReadDirReturnType is the return type for ReadDir, containing the directory entries and an error
type ReadDirReturnType struct {
	Dir []os.DirEntry
	Err error
}

func MockReadDir(fs *FakeFileSystemService, values []ReadDirReturnType) {
	for idx, value := range values {
		fs.ReadDirReturnsOnCall(idx, value.Dir, value.Err)
	}
}

type statFunc func(filename string) (os.FileInfo, error)

// MockFileInfo implements FileInfo, for returning mock stats results
type MockFileInfo struct {
	name    string
	size    int64
	isDir   bool
	mode    os.FileMode
	modTime time.Time
}

// Name returns file name
func (i MockFileInfo) Name() string {
	return i.name
}

// Size returns file size
func (i MockFileInfo) Size() int64 {
	return i.size
}

// Mode returns file mode
func (i MockFileInfo) Mode() os.FileMode {
	return i.mode
}

// ModTime returns file modification time
func (i MockFileInfo) ModTime() time.Time {
	return i.modTime
}

// IsDir returns true if file is a directory
func (i MockFileInfo) IsDir() bool {
	return i.isDir
}

// Sys returns nil, since mocked
func (i MockFileInfo) Sys() any {
	return nil
}

// MockDirEntry implements DirEntry for mocking ReadDir
type MockDirEntry struct {
	FileName string
	FilePath string
	Dir      bool
	FileMode os.FileMode
	StatFunc statFunc
}

// Name returns entry name
func (e MockDirEntry) Name() string {
	return e.FileName
}

// Type returns entry type (aka mode)
func (e MockDirEntry) Type() os.FileMode {
	return e.FileMode
}

// IsDir returns true if entry is a directory
func (e MockDirEntry) IsDir() bool {
	return e.Dir
}

// Info returns entry stat info
func (e MockDirEntry) Info() (os.FileInfo, error) {
	return e.StatFunc(path.Join(e.FilePath, e.FileName))
}
