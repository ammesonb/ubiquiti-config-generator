package github

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/mocks/github_client"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestCheckAuthorization(t *testing.T) {
	svc := DefaultGitHubService{client: &github_client.MockClient{}}
	client.SetNextResult(github_client.CheckAuthorizationFn, []any{nil})
	err := svc.CheckAuthorization(t.Context(), &filesystem.OSService{})
	assert.NoError(t, err)
}
