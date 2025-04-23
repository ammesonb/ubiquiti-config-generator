package github_client

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/google/go-github/v70/github"
)

type githubClientServiceRegistration struct {
}

func (githubClientServiceRegistration) IsSingleton() bool {
	return true
}

func (githubClientServiceRegistration) New() (services.ServiceImplementation, error) {
	return &DefaultGitHubAPIClient{
		client: github.NewClient(nil),
	}, nil
}

func RegisterService(ctx context.Context) error {
	services.RegisterService(services.GitHubClientServiceIndex, githubClientServiceRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.GitHubClientServiceIndex)
	return err
}

func GetService() APIClient {
	svc, _ := services.GetService(services.GitHubClientServiceIndex)
	return svc.(APIClient)
}
