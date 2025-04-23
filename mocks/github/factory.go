package github

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services"
)

type mockGitHubRegistration struct {
}

func (mockGitHubRegistration) IsSingleton() bool {
	return true
}

// New instantiates a new MockGitHubService
func (mockGitHubRegistration) New() (services.ServiceImplementation, error) {
	return &MockGitHubService{}, nil
}

// RegisterService registers the mock GitHub service
func RegisterService(_ context.Context) error {
	services.RegisterService(services.GitHubServiceIndex, mockGitHubRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.GitHubServiceIndex)
	return err
}
