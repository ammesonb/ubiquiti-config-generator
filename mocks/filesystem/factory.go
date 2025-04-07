package filesystem

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/services"
)

type mockFileSystemRegistration struct {
}

// IsSingleton returns filesystem is singleton
func (m mockFileSystemRegistration) IsSingleton() bool {
	return true
}

// New returns a new mock FileSystemService
func (m mockFileSystemRegistration) New() (services.ServiceImplementation, error) {
	return &MockedFileSystem{}, nil
}

// MockFileSystem registers the mocked filesystem service, requiring a test object to ensure only used in tests
func MockFileSystem(_ *testing.T) error {
	services.RegisterService(services.FilesystemServiceIndex, mockFileSystemRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.FilesystemServiceIndex)

	return err
}
