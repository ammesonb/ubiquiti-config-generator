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

var (
	errListingDeployments = "error listing deployments"
	errNoActiveDeployment = "no active deployment found"
	errListingStatus      = "error listing deployment statuses"
)

type APIClient interface {
	SetAuthToken(token string)
	ListInstallations(ctx context.Context) ([]*github.Installation, *github.Response, error)
	CheckAuthorization(ctx context.Context) error
	GetActiveDeployment(owner, repo string, ctx context.Context) (*github.Deployment, error)
	CreateDeployment(owner, repo, beforeSHA, afterSHA string, previousDeployment *github.Deployment, ctx context.Context) (*github.Deployment, error)
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

// SetAuthToken will set the auth token for the client
func (d *DefaultGitHubAPIClient) SetAuthToken(token string) {
	d.client = d.client.WithAuthToken(token)
}

// CheckAuthorization will ensure the client is authorized and valid for at least the next minute
func (d *DefaultGitHubAPIClient) CheckAuthorization(ctx context.Context) error {
	// Ensure valid for at least the next minute, to account for clock skew and request latency
	if d.accessToken == nil || d.accessToken.ExpiresAt.Before(time.Now().Add(time.Minute)) {
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

// ListInstallations will list all installations using the app's access token
func (d *DefaultGitHubAPIClient) ListInstallations(ctx context.Context) ([]*github.Installation, *github.Response, error) {
	return d.client.Apps.ListInstallations(ctx, &github.ListOptions{})
}

// GetAccessToken will return the current access token
func (d *DefaultGitHubAPIClient) GetAccessToken() string {
	return d.accessToken.GetToken()
}

// RevokeAccessToken will revoke the current access token immediately
func (d *DefaultGitHubAPIClient) RevokeAccessToken(ctx context.Context) error {
	_, err := d.client.Apps.RevokeInstallationToken(ctx)
	return err
}

// GetActiveDeployment will return the first active deployment for the given owner and repository
func (d *DefaultGitHubAPIClient) GetActiveDeployment(owner, repo string, ctx context.Context) (*github.Deployment, error) {
	if err := d.CheckAuthorization(ctx); err != nil {
		return nil, err
	}

	// TODO: when multiple devices supported, this needs to be called for each device altered since the modifications
	// may have multiple points of divergence from the base
	iteratorCtx, iteratorCancel := context.WithCancel(context.Background())
	defer iteratorCancel()
	deploymentChan, nextChan, errChan := paginateGitHub(
		func(page int, ctx context.Context) ([]*github.Deployment, *github.Response, error) {
			return d.client.Repositories.ListDeployments(ctx, owner, repo, &github.DeploymentsListOptions{
				ListOptions: github.ListOptions{
					PerPage: 50,
					Page:    page,
				},
			})
		},
		iteratorCtx,
	)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case deployment := <-deploymentChan:
			active, err := d.isDeploymentActive(deployment, owner, repo, ctx)
			if err != nil {
				return nil, errors.ErrWithParent(errListingStatus, err)
			}

			if active {
				return deployment, nil
			}
			nextChan <- true
		case err := <-errChan:
			if err == errStopIteration {
				return nil, errors.Err(errNoActiveDeployment)
			}

			return nil, errors.ErrWithParent(errListingDeployments, err)
		}
	}
}

// isDeploymentActive searches for any status "active" given a deployment, owner and repo
func (d *DefaultGitHubAPIClient) isDeploymentActive(deployment *github.Deployment, owner, repo string, ctx context.Context) (bool, error) {
	paginateCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	statusChan, nextChan, errChan := paginateGitHub(
		func(page int, ctx context.Context) ([]*github.DeploymentStatus, *github.Response, error) {
			return d.client.Repositories.ListDeploymentStatuses(ctx, owner, repo, deployment.GetID(), &github.ListOptions{
				PerPage: 50,
				Page:    page,
			},
			)
		},
		paginateCtx,
	)

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case status := <-statusChan:
			if status.GetState() == "active" {
				return true, nil
			} else {
				nextChan <- true
			}
		case err := <-errChan:
			cancel()
			return false, err
		}
	}
}

// CreateDeployment will create a new deployment between the two SHAs (or provided deployment) for the given owner and repository
func (d *DefaultGitHubAPIClient) CreateDeployment(owner, repo, beforeSHA, afterSHA string, previousDeployment *github.Deployment, ctx context.Context) (*github.Deployment, error) {
	if err := d.CheckAuthorization(ctx); err != nil {
		return nil, err
	}

	previous := beforeSHA
	if previousDeployment != nil {
		previous = (*previousDeployment).GetSHA()
	}
	deployment, _, err := d.client.Repositories.CreateDeployment(ctx, owner, repo, &github.DeploymentRequest{
		Ref: &afterSHA,
		Payload: map[string]interface{}{
			"previous_commit": previous,
		},
	})
	return deployment, err
}
