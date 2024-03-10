package config

import (
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEnumerateConfigFiles(t *testing.T) {
	readName := "read"
	mocks.InitOrClearFuncReturn(readName)

	fsWrap := mocks.FsWrapper{
		ReadDir: func(_ string) ([]os.DirEntry, error) {
			val, err := mocks.GetResult(readName)
			if err != nil {
				return nil, err
			} else if val == "miss" {
				return nil, os.ErrNotExist
				// } else {
				// return {}, nil
			}

			return val.([]os.DirEntry), nil
		},
	}

	assert.NoError(t, mocks.SetNextResult(readName, "miss"))

	config := &DeviceConfig{ConfigFiles: []string{"/config/", "/boot/conf", "/etc/network", "/opt/config/"}}
	files, errs := EnumerateConfigFiles(fsWrap, config, "/")
	assert.Empty(t, files, "No files if read dir fails")
	assert.Len(t, errs, 1, "Does not continue after read fail")
	assert.ErrorIs(t, errs[0], utils.ErrWithCtx(errReadDir, "/"))

	assert.NoError(
		t,
		mocks.SetNextResult(
			readName,
			[]os.DirEntry{
				mocks.MockDirEntry{IName: "boot", IIsDir: true},
				mocks.MockDirEntry{IName: "tmp", IIsDir: true},
				mocks.MockDirEntry{IName: "opt", IIsDir: true},
			},
		),
	)
	assert.NoError(t, mocks.SetNextResult(readName, []os.DirEntry{
		mocks.MockDirEntry{IName: "skipped"},
		mocks.MockDirEntry{IName: "conf"},
	}))
	assert.NoError(t, mocks.SetNextResult(readName, []os.DirEntry{
		mocks.MockDirEntry{IName: "config", IIsDir: true},
		mocks.MockDirEntry{IName: "hostname"},
		mocks.MockDirEntry{IName: "vyatta", IIsDir: true},
	}))
	assert.NoError(t, mocks.SetNextResult(readName, []os.DirEntry{
		mocks.MockDirEntry{IName: "network.yaml"},
		mocks.MockDirEntry{IName: "iface.yaml"},
		mocks.MockDirEntry{IName: "fw.yaml"},
	}))

	files, errs = EnumerateConfigFiles(fsWrap, config, "/")
	assert.Empty(t, errs)
	assert.Len(t, files, 4, "Files found")
	assert.Equal(t, files[0], "/boot/conf", "File match")
	assert.Equal(t, files[1], "/opt/config/network.yaml", "File match")
	assert.Equal(t, files[2], "/opt/config/iface.yaml", "File match")
	assert.Equal(t, files[3], "/opt/config/fw.yaml", "File match")
}

func TestDeviceFilesChanged(t *testing.T) {
	config := &DeviceConfig{
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
