package github

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"
)

type Service interface {
	CheckAuthorization(ctx context.Context, configService configuration.Service, fsService filesystem.Service) error
	SetClient(client github_client.APIClient)
}

type DefaultGitHubService struct {
	client github_client.APIClient
}

func (s *DefaultGitHubService) SetClient(client github_client.APIClient) {
	s.client = client
}

func (s *DefaultGitHubService) CheckAuthorization(ctx context.Context, configService configuration.Service, fsService filesystem.Service) error {
	return s.client.CheckAuthorization(ctx, configService, fsService)
}
