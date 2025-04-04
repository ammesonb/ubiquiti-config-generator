package filesystem

import (
	"os"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestMockFileSystem(t *testing.T) {
	assert.NoError(t, MockFileSystem(t))
	fs := filesystem.GetService().(*MockedFileSystem)
	t.Run("no registered functions", func(t *testing.T) {
		info, err := fs.Stat("service.go")
		assert.Nil(t, info)
		assert.ErrorIs(t, err, errors.ErrWithCtx(mocks.ErrNoSuchFunction, StatFn))

		bytes, err := fs.ReadFile("service.go")
		assert.Nil(t, bytes)
		assert.ErrorIs(t, err, errors.ErrWithCtx(mocks.ErrNoSuchFunction, ReadFileFn))

		entries, err := fs.ReadDir("service.go")
		assert.Nil(t, entries)
		assert.ErrorIs(t, err, errors.ErrWithCtx(mocks.ErrNoSuchFunction, ReadDirFn))

		file, err := fs.Open("service.go")
		assert.Nil(t, file)
		assert.ErrorIs(t, err, errors.ErrWithCtx(mocks.ErrNoSuchFunction, OpenFn))
	})

	fs.Reset()

	t.Run("stat", func(t *testing.T) {
		fileName := "service.go"
		size := int64(1234)
		fs.SetNextResult(StatFn, []any{nil, os.ErrNotExist})
		fs.SetNextResult(StatFn, []any{MockFileInfo{name: fileName, size: size}, nil})

		info, err := fs.Stat(fileName)
		assert.Nil(t, info)
		assert.ErrorIs(t, err, os.ErrNotExist)

		info, err = fs.Stat(fileName)
		assert.NotNil(t, info)
		assert.NoError(t, err)
		assert.Equal(t, fileName, info.Name())
		assert.Equal(t, size, info.Size())
	})

	t.Run("readDir", func(t *testing.T) {
		fs.SetNextResult(ReadDirFn, []any{nil, os.ErrNotExist})
		entryOne := MockDirEntry{FileName: "foo", Dir: false}
		entryTwo := MockDirEntry{FileName: "bar", Dir: true}
		fs.SetNextResult(ReadDirFn, []any{[]MockDirEntry{entryOne, entryTwo}, nil})

		entries, err := fs.ReadDir(".")
		assert.Nil(t, entries)
		assert.ErrorIs(t, err, os.ErrNotExist)

		entries, err = fs.ReadDir(".")
		assert.NoError(t, err)
		assert.Len(t, entries, 2)
		assert.Equal(t, entryOne, entries[0])
		assert.Equal(t, entryTwo, entries[1])
	})

	t.Run("readFile", func(t *testing.T) {
		path := "/var/tmp.txt"
		fs.SetNextResult(ReadFileFn, []any{nil, os.ErrNotExist})
		data := []byte("foo")
		fs.SetNextResult(ReadFileFn, []any{data, nil})

		content, err := fs.ReadFile(path)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Empty(t, content)

		content, err = fs.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, data, content)
	})

	t.Run("open", func(t *testing.T) {
		path := "/var/tmp.txt"
		fs.SetNextResult(OpenFn, []any{nil, os.ErrNotExist})
		file := &os.File{}
		fs.SetNextResult(OpenFn, []any{file, nil})

		f, err := fs.Open(path)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, f)

		f, err = fs.Open(path)
		assert.NoError(t, err)
		assert.Equal(t, file, f)
	})

}
