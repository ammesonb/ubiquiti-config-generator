package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockFileSystemRegistration(t *testing.T) {
	assert.NoError(t, MockFileSystem(t))

	t.Run("singleton", func(t *testing.T) {
		assert.True(t, mockFileSystemRegistration{}.IsSingleton())
	})

	t.Run("new", func(t *testing.T) {
		fs, err := mockFileSystemRegistration{}.New()
		assert.NoError(t, err)
		assert.IsType(t, &MockedFileSystem{}, fs)
	})
}
