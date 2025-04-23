// Services registry to instantiate dependencies used across the application

package services

import (
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/charmbracelet/log"
)

type ServiceImplementation interface {
	StopService(*log.Logger)
}

// ServiceRegistration provides the required functionality to generate/get on-demand features

type ServiceRegistration interface {
	New() (ServiceImplementation, error)
	IsSingleton() bool
}

// ServiceIndex is a unique integer representing a particular registered service
type ServiceIndex string

const (
	ConfigurationServiceIndex ServiceIndex = "configuration"
	FilesystemServiceIndex    ServiceIndex = "filesystem"
	DatabaseServiceIndex      ServiceIndex = "database"
	GitHubServiceIndex        ServiceIndex = "github"
	GitHubClientServiceIndex  ServiceIndex = "github_client"
)

// serviceRegistrations is a map containing all registered services
var serviceRegistrations = make(map[ServiceIndex]ServiceRegistration)

// serviceCache is a map containing all generated instances of services
var serviceCache = make(map[ServiceIndex]ServiceImplementation)

var errNoSuchService = "no such service: %s"

// RegisterService stores a provided service under the given index
func RegisterService(index ServiceIndex, service ServiceRegistration) {
	serviceRegistrations[index] = service
	serviceCache[index] = nil
}

// GetService returns an instance of the service registered under the given index, but prefers to reuse a cached instance
func GetService(index ServiceIndex) (ServiceImplementation, error) {
	registration, ok := serviceRegistrations[index]
	if !ok {
		return nil, errors.ErrWithCtx(errNoSuchService, index)
	}

	// If the service is not a singleton or has not been cached, create a new instance
	if serviceCache[index] == nil || !registration.IsSingleton() {
		return createService(index)
	}

	return serviceCache[index], nil
}

// createService returns a new instance of the service registered under the given index
// To avoid accidentally instantiating extra singletons, this cannot be called externally
func createService(index ServiceIndex) (ServiceImplementation, error) {
	// only called from GetService, so we know the service was registered
	registration := serviceRegistrations[index]
	service, err := registration.New()
	// Only cache the result on success, and for singletons
	if err == nil && registration.IsSingleton() {
		serviceCache[index] = service
	}

	return service, err
}
