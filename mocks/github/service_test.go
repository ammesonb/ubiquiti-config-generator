package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopService(t *testing.T) {
	m := &MockGitHubService{}
	m.StopService(nil)
}

func TestCheckAuthorization(t *testing.T) {
	m := &MockGitHubService{}
	assert.Error(t, m.CheckAuthorization(nil))
	m.SetNextResult(CheckAuthorizationFn, []any{nil})
	assert.NoError(t, m.CheckAuthorization(nil))
}
