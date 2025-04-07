package filesystem

import (
	"os"

	"github.com/ammesonb/ubiquiti-config-generator/mocks"
)

const (
	OpenFn     mocks.FunctionName = "Open"
	ReadFileFn mocks.FunctionName = "ReadFile"
	ReadDirFn  mocks.FunctionName = "ReadDir"
	StatFn     mocks.FunctionName = "Stat"
)

// MockedFileSystem intercepts any calls to the system filesystem and returns customized mocked values
type MockedFileSystem struct {
	*mocks.ServiceMock
}

// Stat returns FileInfo and error
func (m *MockedFileSystem) Stat(path string) (os.FileInfo, error) {
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
func (m *MockedFileSystem) ReadFile(path string) ([]byte, error) {
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
func (m *MockedFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
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
func (m *MockedFileSystem) Open(path string) (*os.File, error) {
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
