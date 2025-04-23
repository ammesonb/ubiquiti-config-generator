package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistration(t *testing.T) {
	registration := githubRegistration{}
	assert.True(t, registration.IsSingleton())
	svc, err := registration.New()
	assert.NoError(t, err)
	assert.IsType(t, &DefaultGitHubService{}, svc)
}

func TestRegisterService(t *testing.T) {
	assert.NoError(t, RegisterService(t.Context()))
}

func TestGetService(t *testing.T) {
	assert.NoError(t, RegisterService(t.Context()))
	svc := GetService()
	assert.IsType(t,
		&DefaultGitHubService{}, svc)
}
