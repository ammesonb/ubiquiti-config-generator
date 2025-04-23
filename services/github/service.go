package github

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"
	"github.com/charmbracelet/log"
)

type GitHubService interface {
	CheckAuthorization(ctx context.Context) error
	SetClient(client github_client.APIClient)
	services.ServiceImplementation
}

type DefaultGitHubService struct {
	client github_client.APIClient
}

// StopService cleans up resources used by the service.
func (s *DefaultGitHubService) StopService(logger *log.Logger) {}

func (s *DefaultGitHubService) SetClient(client github_client.APIClient) {
	s.client = client
}

func (s *DefaultGitHubService) CheckAuthorization(ctx context.Context) error {
	return s.client.CheckAuthorization(ctx)
}
