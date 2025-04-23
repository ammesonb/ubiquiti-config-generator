package config

import (
	"os"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	mockFs "github.com/ammesonb/ubiquiti-config-generator/mocks/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestEnumerateConfigFiles(t *testing.T) {
	assert.NoError(t, mockFs.MockFileSystem(t))
	mockedFs := filesystem.GetService().(*mockFs.MockedFileSystem)
	mockedFs.Reset()
	config := &configuration.DeviceConfig{ConfigFiles: []string{"/config/", "/boot/conf", "/etc/network", "/opt/config/"}}

	t.Run("nonexistent directory", func(t *testing.T) {
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{nil, os.ErrNotExist})
		files, errs := EnumerateConfigFiles(config, "/")
		assert.Nil(t, files, "No files if read dir fails")
		assert.Len(t, errs, 1, "Does not continue after read fail")
		assert.ErrorIs(t, errs[0], errors.ErrWithCtx(errReadDir, "/"))
	})

	t.Run("nested directories considered and errors omitted", func(t *testing.T) {
		mockedFs.ResetFunc(mockFs.ReadDirFn)
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "boot", Dir: true},
				{FileName: "config", Dir: true},
			},
			nil,
		})
		// boot directory is not readable
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{nil, os.ErrPermission})
		// config directory contains one file
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "network.yaml"},
			},
			nil,
		})
		files, errs := EnumerateConfigFiles(config, "/")
		assert.Len(t, files, 1, "Only one file in boot directory")
		assert.Equal(t, files[0], "/config/network.yaml")
		assert.Len(t, errs, 1, "One error from reading boot")
		assert.ErrorIs(t, errs[0], errors.ErrWithCtx(errReadDir, "/boot"))
	})

	t.Run("mix of skipped and found files", func(t *testing.T) {
		mockedFs.ResetFunc(mockFs.ReadDirFn)
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "boot", Dir: true},
				{FileName: "etc", Dir: true},
				{FileName: "opt", Dir: true},
				{FileName: "tmp", Dir: true},
			},
			nil,
		})
		// boot contains a conf directory and grub file
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "conf", Dir: true},
				{FileName: "grub.conf"},
			},
			nil,
		})

		// conf directory contains one file, but skipped since conf in config is a file
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "boot.yaml"},
			},
			nil,
		})

		// etc directory contains various files
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "network"},
				{FileName: "interface"},
				{FileName: "firewall"},
			},
			nil,
		})

		// opt directory contains a config directory
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "config", Dir: true},
			},
			nil,
		})

		// config directory contains two files
		mockedFs.SetNextResult(mockFs.ReadDirFn, []any{
			[]mockFs.MockDirEntry{
				{FileName: "network.conf"},
				{FileName: "interface.conf"},
			},
			nil,
		})

		files, errs := EnumerateConfigFiles(config, "/")

		assert.Len(t, files, 3)
		if len(files) < 3 {
			t.FailNow()
		}
		assert.Equal(t, "/etc/network", files[0])
		assert.Equal(t, "/opt/config/network.conf", files[1])
		assert.Equal(t, "/opt/config/interface.conf", files[2])
		assert.Empty(t, errs, "No errors")
	})
}

func TestDeviceFilesChanged(t *testing.T) {
	config := &configuration.DeviceConfig{
		ConfigFiles: []string{
			"/foo",
			"/bar",
			"/baz/",
			"/ipsum/lorem",
			"/ipsum/lorem2/",
		},
	}

	assert.False(t, DeviceFilesChanged(config, []string{}), "Empty list not changed")
	assert.False(t, DeviceFilesChanged(config, []string{"/lorem"}), "Unchanged")
	assert.False(t, DeviceFilesChanged(config, []string{"./ipsum"}), "Local file, not root unchanged")
	assert.False(
		t,
		DeviceFilesChanged(config, []string{"/ipsum"}),
		"Parent directory is not included just because child is",
	)
	assert.False(t, DeviceFilesChanged(config, []string{"/foo2"}), "File with same prefix is not included")
	assert.False(
		t,
		DeviceFilesChanged(config, []string{"/ipsum/lorem/file"}),
		"File with similar parent directory not included",
	)

	assert.True(t, DeviceFilesChanged(config, []string{"/foo"}))
	assert.True(t, DeviceFilesChanged(config, []string{"/baz/file"}))
	assert.True(t, DeviceFilesChanged(config, []string{"/ipsum/lorem"}))
	assert.True(t, DeviceFilesChanged(config, []string{"/ipsum/lorem2/something"}))
}

func TestDirInConfig(t *testing.T) {
	files := []string{"/config/boot", "/opt/config/"}

	assert.False(t, dirInConfig(files, "/boot"))
	assert.False(t, dirInConfig(files, "/boot/"))
	assert.False(t, dirInConfig(files, "/config/boot/"))
	assert.True(t, dirInConfig(files, "/opt"))
	assert.True(t, dirInConfig(files, "/opt/"))
	assert.True(t, dirInConfig(files, "/opt/config/"))
	assert.True(t, dirInConfig(files, "/config/"))
}
