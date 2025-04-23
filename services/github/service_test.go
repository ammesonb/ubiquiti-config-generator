package github

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/mocks/github_client"
	"github.com/stretchr/testify/assert"
)

func TestCheckAuthorization(t *testing.T) {
	assert.NoError(t, RegisterService(t.Context()))
	svc := GetService()
	client := github_client.MockClient{}
	svc.SetClient(&client)
	client.SetNextResult(github_client.CheckAuthorizationFn, []any{nil})
	err := svc.CheckAuthorization(t.Context())
	assert.NoError(t, err)
}
