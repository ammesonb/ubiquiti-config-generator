package github

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"
	"github.com/google/go-github/v70/github"
)

type Service interface {
	SetClient(client github_client.APIClient)
	CreateDeployment(owner, repo, beforeSHA, afterSHA string, ctx context.Context) (*github.Deployment, string, string, error)
}

type defaultGitHubService struct {
	client github_client.APIClient
}

func (s *defaultGitHubService) SetClient(client github_client.APIClient) {
	s.client = client
}

func (s *defaultGitHubService) CreateDeployment(owner, repo, beforeSHA, afterSHA string, ctx context.Context) (*github.Deployment, string, string, error) {
	previous, err := s.client.GetActiveDeployment(owner, repo, ctx)
	if err != nil {
		return nil, "", "", err
	}

	deployment, err := s.client.CreateDeployment(owner, repo, beforeSHA, afterSHA, previous, ctx)
	if err != nil {
		return nil, "", "", err
	}

	return deployment, beforeSHA, afterSHA, nil
}

// New creates a new GitHub service instance
func New(client github_client.APIClient, ctx context.Context) (Service, error) {
	if err := client.CheckAuthorization(ctx); err != nil {
		return nil, err
	}
	return &defaultGitHubService{
		client: client,
	}, nil
}
