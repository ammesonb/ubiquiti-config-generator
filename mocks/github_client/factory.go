package github_client

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services"
)

type mockGitHubClientRegistration struct{}

func (mockGitHubClientRegistration) IsSingleton() bool {
	return true
}
func (mockGitHubClientRegistration) New() (services.ServiceImplementation, error) {
	return &MockClient{}, nil
}

func RegisterService(ctx context.Context) error {
	services.RegisterService(services.GitHubClientServiceIndex, mockGitHubClientRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.GitHubClientServiceIndex)
	return err
}
