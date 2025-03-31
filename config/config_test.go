package config

import (
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/internal/test_helpers"
	"testing"

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
func TestReadConfig(t *testing.T) {
	content, err := ReadConfig("./nonexistent")
	assert.NotNil(t, err, "Error for nonexistent path should not be nil")
	assert.Empty(t, content, "No content for nonexistent file")

	content, err = ReadConfig("./config.go")
	assert.Nil(t, err, "No error reading file")
	assert.NotEmpty(t, content, "Content returned from file")
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig([]byte("invalid[yaml"))
	assert.NotNil(t, err, "Invalid YAML should throw error")
	assert.ErrorIs(t, err, errors.Err(errFailParseConfig))
	assert.Nil(t, config, "Invalid YAML should not return config")

	config, err = LoadConfig([]byte("{}"))
	assert.Nil(t, err, "No error loading empty config")
	assert.NotNil(t, config, "Empty config is non-nil")
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
