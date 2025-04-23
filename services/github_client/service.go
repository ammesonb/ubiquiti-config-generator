package github_client

import (
	"context"
	"slices"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/google/go-github/v70/github"
	gh "github.com/google/go-github/v70/github"
)

type APIClient interface {
	WithAuthToken(token string)
	ListInstallations(ctx context.Context) ([]*gh.Installation, *gh.Response, error)
	CheckAuthorization(ctx context.Context, configService configuration.Service, fsService filesystem.Service) error
}

type DefaultGitHubAPIClient struct {
	client      *gh.Client
	accessToken *gh.InstallationToken
}

func (d *DefaultGitHubAPIClient) WithAuthToken(token string) {
	d.client = d.client.WithAuthToken(token)
}

func (d *DefaultGitHubAPIClient) CheckAuthorization(ctx context.Context, configService configuration.Service, fsService filesystem.Service) error {
	if d.accessToken == nil || d.accessToken.ExpiresAt.Before(time.Now()) {
		return d.authorize(ctx, configService, fsService)
	}
	return nil
}

func (d *DefaultGitHubAPIClient) authorize(ctx context.Context, configService configuration.Service, fsService filesystem.Service) error {
	gitConfig := configService.GetGitConfig()
	jwt, err := makeJWT(gitConfig, fsService)
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

func (d *DefaultGitHubAPIClient) getAppInstallation(ctx context.Context, jwt string, appId string) (*gh.Installation, error) {
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

func (d *DefaultGitHubAPIClient) ListInstallations(ctx context.Context, options *gh.ListOptions) ([]*gh.Installation, *gh.Response, error) {
	return d.client.Apps.ListInstallations(ctx, options)
}
