package configuration

import (
	"os"
	"testing"

	mockFs "github.com/ammesonb/ubiquiti-config-generator/mocks/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/internal/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestShouldGetEnv(t *testing.T) {
	assert.False(t, shouldGetValueFromEnv("some_value"))
	assert.False(t, shouldGetValueFromEnv("special$#!@characters^&*"))
	assert.True(t, shouldGetValueFromEnv("$SOME_ENV_VAR"))
}

func TestTrimYAMLEnv(t *testing.T) {
	assert.Equal(t, "value", stripEnvMarker("value"))
	assert.Equal(t, "value", stripEnvMarker("$value"))
	assert.Equal(t, "value$", stripEnvMarker("$value$"))
}

func TestUpdateConfigFromEnv(t *testing.T) {
	// Set test environment values and reset after test
	envValues := map[string]string{
		"addr":           "1.2.3.4",
		"keyfile":        "/etc/keyfile",
		"log_db":         "/etc/log.db",
		"log_user":       "log_user",
		"log_pass":       "log_pass",
		"git_key":        "/etc/git_kf",
		"webhook":        "http://example.com",
		"listen_ip":      "0.0.0.0:8080",
		"webhook_secret": "abcdef",
	}

	for k, v := range envValues {
		t.Setenv(k, v)
	}

	t.Run("explicit values are not overwritten", func(t *testing.T) {
		config := Config{
			Logging: LoggingConfig{
				DBName:   "logs.db",
				User:     "log_user",
				Password: "password",
			},
			Git: GitConfig{
				PrivateKeyPath: "/keyfile",
				WebhookURL:     "localhost",
				ListenIP:       "localhost",
				WebhookSecret:  "secret",
			},
		}

		getConfigValuesFromEnv(&config)

		tracker := test_helpers.NewAssertionTracker(t)
		// Need explicit strings since config could be changed
		tracker.Expect("Logging.DBName", "logs.db", config.Logging.DBName)
		tracker.Expect("Logging.User", "log_user", config.Logging.User)
		tracker.Expect("Logging.Password", "password", config.Logging.Password)
		tracker.Expect("Git.PrivateKeyPath", "/keyfile", config.Git.PrivateKeyPath)
		tracker.Expect("Git.WebhookURL", "localhost", config.Git.WebhookURL)
		tracker.Expect("Git.ListenIP", "localhost", config.Git.ListenIP)
		tracker.Expect("Git.WebhookSecret", "secret", config.Git.WebhookSecret)
	})

	t.Run("environment values are loaded", func(t *testing.T) {
		config := Config{
			Logging: LoggingConfig{
				DBName:   "$log_db",
				User:     "your_user",
				Password: "$log_pass",
			},
			Git: GitConfig{
				PrivateKeyPath: "$git_key",
				WebhookURL:     "$webhook",
				ListenIP:       "$listen_ip",
				WebhookSecret:  "$webhook_secret",
			},
		}

		getConfigValuesFromEnv(&config)

		tracker := test_helpers.NewAssertionTracker(t)
		tracker.Expect("Logging.DBName", "/etc/log.db", config.Logging.DBName)
		tracker.Expect("Logging.User", "your_user", config.Logging.User)
		tracker.Expect("Logging.Password", "log_pass", config.Logging.Password)
		tracker.Expect("Git.PrivateKeyPath", "/etc/git_kf", config.Git.PrivateKeyPath)
		tracker.Expect("Git.WebhookURL", "http://example.com", config.Git.WebhookURL)
		tracker.Expect("Git.ListenIP", "0.0.0.0:8080", config.Git.ListenIP)
		tracker.Expect("Git.WebhookSecret", "abcdef", config.Git.WebhookSecret)
	})
}

func TestServiceGetters(t *testing.T) {
	gitConfig := GitConfig{AppID: "12345"}
	logConfig := LoggingConfig{
		DBName: "log.db",
	}
	var devices []*DeviceConfig

	s := DefaultConfigurationService{
		config: &Config{
			Git:     gitConfig,
			Logging: logConfig,
			Devices: devices,
		},
	}

	assert.Equal(t, gitConfig, s.GetGitConfig())
	assert.Equal(t, logConfig, s.GetLoggingConfig())
	assert.Equal(t, devices, s.GetDevices())
}

func TestServiceLoad(t *testing.T) {
	s := DefaultConfigurationService{}
	assert.NoError(t, mockFs.MockFileSystem(t))
	mockedFs := filesystem.GetService().(*mockFs.MockedFileSystem)
	mockedFs.Reset()

	t.Run("file does not exist", func(t *testing.T) {
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{nil, os.ErrNotExist})
		err := s.Load("/does/not/exist")
		assert.ErrorIs(t, err, errors.ErrWithParent(errReadMainConfig, os.ErrNotExist))
	})

	t.Run("invalid yaml", func(t *testing.T) {
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{[]byte("invalid yaml"), nil})
		yamlFile := "./test-files/invalid-yaml"
		err := s.Load(yamlFile)
		assert.ErrorIs(t, err, errors.ErrWithCtx(errParseMainConfig, yamlFile))
	})

	t.Run("valid config", func(t *testing.T) {
		validConfig := []byte(`
# Interface settings
logging:
  user: $UBQ_CONFIG_USER
  pass: $UBQ_CONFIG_PASS

git:
  # The ID of the application
  app-id: $GITHUB_APP_ID
  # The main GitHub branch
  primary-branch: main

  private-key-path: ./ubiquiti-config.pem
  # The address for the webserver to listen on
  listen-ip: 0.0.0.0
  # The external URL of the webhook, should match the GitHub app configuration
  webhook-url: https://example.com
  # The port to listen on for GitHub webhook stuff
  webhook-port: 12345
  # The secret to use with GitHub webhooks
  webhook-secret: $UBQ_GITHUB_WEBHOOK_SECRET
`)
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{validConfig, nil})
	})

	// TODO: more tests here
}

func TestLoadDevices(t *testing.T) {
	assert.NoError(t, mockFs.MockFileSystem(t))
	mockedFs := filesystem.GetService().(*mockFs.MockedFileSystem)
	mockedFs.Reset()

	t.Run("nonexistent file", func(t *testing.T) {
		mockedFs.ResetFunc(mockFs.ReadFileFn)
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{nil, os.ErrNotExist})
		err := loadDevices("/does/not/exist", &[]*DeviceConfig{})
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		mockedFs.ResetFunc(mockFs.ReadFileFn)
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{[]byte("invalid yaml"), nil})
		var devices []*DeviceConfig
		err := loadDevices("/does/not/exist", &devices)
		assert.ErrorIs(t, err, errors.ErrWithCtx(errParseDevices, "/does/not/exist"))
	})

	t.Run("Explicit device values are not overwritten", func(t *testing.T) {
		deviceYAML := []byte(`
- name: dev1
  address: 1.2.3.4
  keyfile: /etc/keyfile
- name: dev2
  address: 5.6.7.8
`)
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{deviceYAML, nil})
		devices := []*DeviceConfig{}
		err := loadDevices("explicit-values.yaml", &devices)
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
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{deviceYAML, nil})
		devices := []*DeviceConfig{}
		err := loadDevices("env-values.yaml", &devices)
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
