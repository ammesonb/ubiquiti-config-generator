package filesystem

import "os"

type statFunc func(filename string) (os.FileInfo, error)

// FileSystemService provides various ways to access the file system
type FileSystemService interface {
	Stat(filename string) (os.FileInfo, error)
	ReadFile(filename string) ([]byte, error)
	ReadDir(dir string) ([]os.DirEntry, error)
	Open(filename string) (*os.File, error)
}

// DefaultFileSystemService provides default implementations for FileSystemService
type DefaultFileSystemService struct{}

// Stat calls os.Stat
func (s DefaultFileSystemService) Stat(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// ReadFile calls os.ReadFile
func (s DefaultFileSystemService) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// ReadDir calls os.ReadDir
func (s DefaultFileSystemService) ReadDir(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}

// Open calls os.Open
func (s DefaultFileSystemService) Open(filename string) (*os.File, error) {
	return os.Open(filename)
}
