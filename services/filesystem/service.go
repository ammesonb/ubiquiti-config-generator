package filesystem

import (
	"os"

	"github.com/charmbracelet/log"
)

//go:generate go tool counterfeiter -generate

// Service provides various ways to access the file system
//
//counterfeiter:generate . Service
type Service interface {
	Stat(filename string) (os.FileInfo, error)
	ReadFile(filename string) ([]byte, error)
	ReadDir(dir string) ([]os.DirEntry, error)
	Open(filename string) (*os.File, error)
}

// OSService provides a default implementations for FileSystemService using the os library
type OSService struct{}

// Stat calls os.Stat
func (s *OSService) Stat(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// ReadFile calls os.ReadFile
func (s *OSService) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// ReadDir calls os.ReadDir
func (s *OSService) ReadDir(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}

// Open calls os.Open
func (s *OSService) Open(filename string) (*os.File, error) {
	return os.Open(filename)
}

func (s *OSService) StopService(_ *log.Logger) {}

func New() (Service, error) {
	return &OSService{}, nil
}
