package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLiveFileSystemService(t *testing.T) {
	fs, err := New()
	assert.NoError(t, err)

	t.Run("stat", func(t *testing.T) {
		info, err := fs.Stat("service.go")
		assert.NoError(t, err)

		assert.Greater(t, info.Size(), int64(0))
		assert.False(t, info.IsDir())
	})

	t.Run("readDir", func(t *testing.T) {
		entries, err := fs.ReadDir(".")
		assert.NoError(t, err)

		assert.Greater(t, len(entries), 0)
		foundService := false
		for _, entry := range entries {
			assert.False(t, entry.IsDir() && entry.Name() != "filesystemfakes")
			foundService = foundService || entry.Name() == "service.go"
		}
		assert.True(t, foundService)
	})

	t.Run("readFile", func(t *testing.T) {
		data, err := fs.ReadFile("service.go")
		assert.NoError(t, err)

		assert.Greater(t, len(data), 0)
		assert.Contains(t, string(data), "FileSystemService")
	})

	t.Run("open", func(t *testing.T) {
		file, err := fs.Open("service.go")
		assert.NoError(t, err)
		defer func() { _ = file.Close() }()
		assert.NotNil(t, file)
	})
}
