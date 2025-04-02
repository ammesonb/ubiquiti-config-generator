package filesystem

import (
	"os"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/services"
)

const (
	OpenFn     mocks.FunctionName = "Open"
	ReadFileFn mocks.FunctionName = "ReadFile"
	ReadDirFn  mocks.FunctionName = "ReadDir"
	StatFn     mocks.FunctionName = "Stat"
)

// MockedFileSystem intercepts any calls to the system filesystem and returns customized mocked values
type MockedFileSystem struct {
	mocks.FunctionMock
}

// Stat returns FileInfo and error
func (m MockedFileSystem) Stat(path string) (os.FileInfo, error) {
	values, err := m.GetResult(StatFn, path)
	if err != nil {
		return nil, err
	}

	err = mocks.AnyToError(values[1])
	if err != nil {
		return nil, err
	}

	return values[0].(os.FileInfo), nil
}

// ReadFile returns bytes and error
func (m MockedFileSystem) ReadFile(path string) ([]byte, error) {
	values, err := m.GetResult(ReadFileFn, path)
	if err != nil {
		return nil, err
	}

	err = mocks.AnyToError(values[1])
	if err != nil {
		return nil, err
	}
	return values[0].([]byte), nil
}

// ReadDir returns DirEntry and error
func (m MockedFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	values, err := m.GetResult(ReadDirFn, path)
	if err != nil {
		return nil, err
	}

	err = mocks.AnyToError(values[1])
	if err != nil {
		return nil, err
	}
	// cannot directly cast a slice, so instead convert each entry separately
	var entries []os.DirEntry
	for _, entry := range values[0].([]MockDirEntry) {
		entries = append(entries, entry)
	}
	return entries, nil
}

// Open returns File and error
func (m MockedFileSystem) Open(path string) (*os.File, error) {
	values, err := m.GetResult(OpenFn, path)
	if err != nil {
		return nil, err
	}

	err = mocks.AnyToError(values[1])
	if err != nil {
		return nil, err
	}
	return values[0].(*os.File), nil
}

type mockFileSystemRegistration struct {
}

// IsSingleton returns filesystem is not singleton
func (m mockFileSystemRegistration) IsSingleton() bool {
	return false
}

// New returns a new mock FileSystemService
func (m mockFileSystemRegistration) New() (any, error) {
	return MockedFileSystem{}, nil
}

// MockFileSystem registers the mocked filesystem service, requiring a test object to ensure only used in tests
func MockFileSystem(_ *testing.T) error {
	services.RegisterService(services.FilesystemServiceIndex, mockFileSystemRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.FilesystemServiceIndex)

	return err
}
