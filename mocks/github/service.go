package github

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/services/github"
	gh "github.com/google/go-github/v70/github"
)

var (
	CheckAuthorizationFn mocks.FunctionName = "CheckAuthorization"
)

type MockGitHubService struct {
	mocks.ServiceMock
	github.GitHubService
}

func (m *MockGitHubService) StopService(_ *log.Logger) {}

func (m *MockGitHubService) CheckAuthorization(_ context.Context) error {
	result, err := m.GetResult(CheckAuthorizationFn)
	if err != nil {
		return err
	}
	if result[0] != nil {
		return result[0].(error)
	}
	return nil
}

func (m *MockGitHubService) SetClient(_ *gh.Client) {
}
