package filesystem

import (
	"github.com/ammesonb/ubiquiti-config-generator/services"
)

type filesystemRegistration struct {
}

// IsSingleton returns filesystem is singleton
func (filesystemRegistration) IsSingleton() bool {
	return true
}

// New instantiates a new FileSystemService
func (filesystemRegistration) New() (services.ServiceImplementation, error) {
	return &DefaultFileSystemService{}, nil
}

// RegisterService registers the filesystem service
func RegisterService() error {
	services.RegisterService(services.FilesystemServiceIndex, filesystemRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.FilesystemServiceIndex)
	return err
}

// GetService returns the filesystem service
func GetService() FileSystemService {
	svc, _ := services.GetService(services.FilesystemServiceIndex)
	return svc.(FileSystemService)
}
