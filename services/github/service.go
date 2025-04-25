package github

import (
	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"
)

type Service interface {
	SetClient(client github_client.APIClient)
}

type DefaultGitHubService struct {
	client github_client.APIClient
}

func (s *DefaultGitHubService) SetClient(client github_client.APIClient) {
	s.client = client
}
