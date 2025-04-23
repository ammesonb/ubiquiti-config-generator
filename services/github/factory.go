package github

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"
)

type githubRegistration struct {
}

// IsSingleton returns github is singleton
func (githubRegistration) IsSingleton() bool {
	return true
}

// New instantiates a new GitHubService
func (githubRegistration) New() (services.ServiceImplementation, error) {
	return &DefaultGitHubService{
		client: github_client.GetService(),
	}, nil
}

// RegisterService registers the github service
func RegisterService(ctx context.Context) error {
	services.RegisterService(services.GitHubServiceIndex, githubRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.GitHubServiceIndex)
	return err
}

// GetService returns the github service
func GetService() GitHubService {
	svc, _ := services.GetService(services.GitHubServiceIndex)
	return svc.(GitHubService)
}
