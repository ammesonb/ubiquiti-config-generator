package config

import (
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/internal/test_helpers"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetDeviceConfigs(t *testing.T) {
	t.Run("Invalid YAML", func(t *testing.T) {
		devices, err := GetDeviceConfigs([]byte("invalid[yaml"))
		assert.Empty(t, devices)
		assert.Error(t, err)
	})

	t.Run("Explicit device values are not overwritten", func(t *testing.T) {
		deviceYAML := []byte(`
- name: dev1
  address: 1.2.3.4
  keyfile: /etc/keyfile
- name: dev2
  address: 5.6.7.8
`)
		devices, err := GetDeviceConfigs(deviceYAML)
		assert.NoError(t, err)

		if len(devices) != 2 {
			t.Fatalf("expected 2 devices, got %d", len(devices))
		}

		tracker := test_helpers.NewAssertionTracker(t)
		tracker.Expect("dev1 name", "dev1", devices[0].Name)
		tracker.Expect("dev1 address", "1.2.3.4", devices[0].Address)
		tracker.Expect("dev1 keyfile", "/etc/keyfile", devices[0].KeyFile)
		tracker.Expect("dev2 name", "dev2", devices[1].Name)
		tracker.Expect("dev2 address", "5.6.7.8", devices[1].Address)
		tracker.Expect("dev2 keyfile", "", devices[1].KeyFile)
	})

	t.Run("Environment values are loaded", func(t *testing.T) {
		envValues := map[string]string{
			"dev1_address": "1.2.3.4",
			"dev1_keyfile": "/etc/keyfile",
			"dev2_address": "5.6.7.8",
		}

		for k, v := range envValues {
			t.Setenv(k, v)
		}

		deviceYAML := []byte(`
- name: dev1
  address: $dev1_address
  keyfile: $dev1_keyfile
- name: dev2
  address: $dev2_address
  port: 80
`)
		devices, err := GetDeviceConfigs(deviceYAML)
		assert.NoError(t, err)

		if len(devices) != 2 {
			t.Fatalf("expected 2 devices, got %d", len(devices))
		}

		tracker := test_helpers.NewAssertionTracker(t)
		tracker.Expect("dev1 name", "dev1", devices[0].Name)
		tracker.Expect("dev1 address", envValues["dev1_address"], devices[0].Address)
		tracker.Expect("dev1 keyfile", envValues["dev1_keyfile"], devices[0].KeyFile)
		tracker.Expect("dev2 name", "dev2", devices[1].Name)
		tracker.Expect("dev2 address", envValues["dev2_address"], devices[1].Address)
		tracker.Expect("dev2 keyfile", "", devices[1].KeyFile)
		tracker.Expect("dev2 keyfile", "80", devices[1].Port)
	})
}

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
	assert.ErrorIs(t, errs[0], errors.ErrWithCtx(errReadDir, "/"))

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
