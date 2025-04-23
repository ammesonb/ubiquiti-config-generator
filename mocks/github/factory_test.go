package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRegistration(t *testing.T) {
	registration := mockGitHubRegistration{}
	assert.True(t, registration.IsSingleton())
	mock, err := registration.New()
	assert.NoError(t, err)
	assert.IsType(t, &MockGitHubService{}, mock)
}

func TestRegisterService(t *testing.T) {
	assert.NoError(t, RegisterService(t.Context()))
}
