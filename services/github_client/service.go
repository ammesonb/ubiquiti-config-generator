package github_client

import (
	"context"
	"slices"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/google/go-github/v70/github"
)

type APIClient interface {
	SetAuthToken(token string)
	ListInstallations(ctx context.Context) ([]*github.Installation, *github.Response, error)
	CheckAuthorization(ctx context.Context) error
	GetAccessToken() string
	RevokeAccessToken(ctx context.Context) error
}

type DefaultGitHubAPIClient struct {
	configService configuration.Service
	fsService     filesystem.Service
	client        *github.Client
	accessToken   *github.InstallationToken
}

func NewAPIClient(configService configuration.Service, fsService filesystem.Service) APIClient {
	return &DefaultGitHubAPIClient{
		configService: configService,
		fsService:     fsService,
		client:        github.NewClient(nil),
	}
}

func (d *DefaultGitHubAPIClient) SetAuthToken(token string) {
	d.client = d.client.WithAuthToken(token)
}

func (d *DefaultGitHubAPIClient) CheckAuthorization(ctx context.Context) error {
	if d.accessToken == nil || d.accessToken.ExpiresAt.Before(time.Now()) {
		return d.authorize(ctx)
	}
	return nil
}

func (d *DefaultGitHubAPIClient) authorize(ctx context.Context) error {
	gitConfig := d.configService.GetGitConfig()
	jwt, err := makeJWT(gitConfig, d.fsService)
	if err != nil {
		return err
	}

	installation, err := d.getAppInstallation(
		ctx,
		jwt,
		gitConfig.AppID,
	)
	if err != nil {
		return err
	}

	d.accessToken, _, err = d.client.Apps.CreateInstallationToken(ctx, installation.GetID(), &github.InstallationTokenOptions{})
	if err != nil {
		return err
	}

	d.client = d.client.WithAuthToken(d.accessToken.GetToken())
	return nil
}

func (d *DefaultGitHubAPIClient) getAppInstallation(ctx context.Context, jwt string, appId string) (*github.Installation, error) {
	installs, _, err := d.client.WithAuthToken(jwt).Apps.ListInstallations(
		ctx,
		&github.ListOptions{},
	)

	if err != nil {
		return nil, errors.ErrWithParent(errListingInstallations, err)
	} else if len(installs) == 0 {
		return nil, errors.Err(errNoInstallations)
	}

	installIndex := slices.IndexFunc(installs, installMatchesAppID(appId))
	if installIndex < 0 {
		return nil, errors.Err(errNoAppIDMatch)
	}

	return installs[installIndex], nil
}

func (d *DefaultGitHubAPIClient) ListInstallations(ctx context.Context) ([]*github.Installation, *github.Response, error) {
	return d.client.Apps.ListInstallations(ctx, &github.ListOptions{})
}

func (d *DefaultGitHubAPIClient) GetAccessToken() string {
	return d.accessToken.GetToken()
}

func (d *DefaultGitHubAPIClient) RevokeAccessToken(ctx context.Context) error {
	_, err := d.client.Apps.RevokeInstallationToken(ctx)
	return err
}
