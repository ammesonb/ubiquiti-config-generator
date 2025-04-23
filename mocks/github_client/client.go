package github_client

import (
	"context"

	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/google/go-github/v70/github"
)

type MockClient struct {
	mocks.ServiceMock
	github.Client
}

const (
	ListInstallationsFn  mocks.FunctionName = "ListInstallations"
	CheckAuthorizationFn mocks.FunctionName = "CheckAuthorization"
)

// ListInstallations mocks the GitHub Apps.ListInstallations method
func (m *MockClient) ListInstallations(ctx context.Context) ([]*github.Installation, *github.Response, error) {
	values, err := m.GetResult(ListInstallationsFn, ctx)
	if err != nil {
		return nil, nil, err
	}

	err = mocks.AnyToError(values[2])
	if err != nil {
		return nil, nil, err
	}

	return values[0].([]*github.Installation), nil, nil
}

// CheckAuthorization mocks the GitHub Apps.CheckAuthorization method
func (m *MockClient) CheckAuthorization(ctx context.Context, _ filesystem.Service) error {
	values, err := m.GetResult(CheckAuthorizationFn, ctx)
	if err != nil {
		return err
	}
	return mocks.AnyToError(values[0])
}

// WithAuthToken returns the mock client, no-op since we don't want to use the actual client
func (m *MockClient) WithAuthToken(token string) {
}
