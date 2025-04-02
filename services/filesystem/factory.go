package filesystem

import "github.com/ammesonb/ubiquiti-config-generator/services"

type filesystemRegistration struct {
}

// IsSingleton returns filesystem is not singleton
func (r filesystemRegistration) IsSingleton() bool {
	return false
}

// New returns a new FileSystemService
func (r filesystemRegistration) New() (any, error) {
	return DefaultFileSystemService{}, nil
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
