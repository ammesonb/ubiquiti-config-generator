package github

import (
	"context"
	"net/http"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"
	"github.com/charmbracelet/log"
)

type accessTokenContext string

var accessTokenKey accessTokenContext = "github_access_token"

type authMiddleware struct {
	configService configuration.Service
	logger        *log.Logger
	client        github_client.APIClient
}

func (m *authMiddleware) AccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := m.client.CheckAuthorization(r.Context())
		if err != nil {
			m.logger.Errorf("Failed checking authorization: %v", err)
			return
		}

		next.ServeHTTP(
			w,
			r.WithContext(
				context.WithValue(r.Context(), accessTokenKey, m.client.GetAccessToken()),
			),
		)

		if err = m.client.RevokeAccessToken(r.Context()); err != nil {
			m.logger.Errorf("Failed revoking access token: %v", err)
		}
	})
}

func NewAuthMiddleware(configService configuration.Service, logger *log.Logger, client github_client.APIClient) *authMiddleware {
	return &authMiddleware{
		configService: configService,
		logger:        logger,
		client:        client,
	}
}
