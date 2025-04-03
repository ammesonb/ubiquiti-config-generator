package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetService(t *testing.T) {
	assert.NoError(t, RegisterService())
	fsOne := GetService()
	fsTwo := GetService()

	assert.NotNil(t, fsOne)
	assert.NotNil(t, fsTwo)
	assert.IsType(t, &DefaultFileSystemService{}, fsOne)
	assert.Equal(t, fsOne, fsTwo, "Same service interface returned")
	assert.True(t, fsOne == fsTwo, "Same service instance returned")
}

func TestRegisterService(t *testing.T) {
	assert.NoError(t, RegisterService())
}
