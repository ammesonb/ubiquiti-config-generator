package config

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldGetEnv(t *testing.T) {
	assert.False(t, shouldGetEnv("some_value"))
	assert.False(t, shouldGetEnv("special$#!@characters^&*"))
	assert.True(t, shouldGetEnv("$SOME_ENV_VAR"))
}

func TestTrimYAMLEnv(t *testing.T) {
	assert.Equal(t, "value", trimYamlEnv("value"))
	assert.Equal(t, "value", trimYamlEnv("$value"))
	assert.Equal(t, "value$", trimYamlEnv("$value$"))
}

func TestConvertEnv(t *testing.T) {
	assert.Nil(t, os.Setenv("addr", "address"))
	assert.Nil(t, os.Setenv("user", "root"))
	assert.Nil(t, os.Setenv("keyfile", "/etc/kf"))
	assert.Nil(t, os.Setenv("log_db", "/etc/log.db"))
	assert.Nil(t, os.Setenv("log_user", "log_root"))
	assert.Nil(t, os.Setenv("log_pass", "log_pass"))
	assert.Nil(t, os.Setenv("git_key", "/etc/git_kf"))
	assert.Nil(t, os.Setenv("webhook", "http://example.com"))
	assert.Nil(t, os.Setenv("listen_ip", "0.0.0.0:8080"))
	assert.Nil(t, os.Setenv("webhook_secret", "abcdef"))

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

	convertEnv(&config)

	assert.Equal(t, "logs.db", config.Logging.DBName)
	assert.Equal(t, "log_user", config.Logging.User)
	assert.Equal(t, "password", config.Logging.Password)
	assert.Equal(t, "/keyfile", config.Git.PrivateKeyPath)
	assert.Equal(t, "localhost", config.Git.WebhookURL)
	assert.Equal(t, "localhost", config.Git.ListenIP)
	assert.Equal(t, "secret", config.Git.WebhookSecret)

	config = Config{
		Logging: LoggingConfig{
			DBName:   "$log_db",
			User:     "$log_user",
			Password: "$log_pass",
		},
		Git: GitConfig{
			PrivateKeyPath: "$git_key",
			WebhookURL:     "$webhook",
			ListenIP:       "$listen_ip",
			WebhookSecret:  "$webhook_secret",
		},
	}

	convertEnv(&config)

	assert.Equal(t, "/etc/log.db", config.Logging.DBName)
	assert.Equal(t, "log_root", config.Logging.User)
	assert.Equal(t, "log_pass", config.Logging.Password)
	assert.Equal(t, "/etc/git_kf", config.Git.PrivateKeyPath)
	assert.Equal(t, "http://example.com", config.Git.WebhookURL)
	assert.Equal(t, "0.0.0.0:8080", config.Git.ListenIP)
	assert.Equal(t, "abcdef", config.Git.WebhookSecret)
}

func TestConvertDeviceEnv(t *testing.T) {
	devs := map[string]*DeviceConfig{
		"device": {
			Address:  "1.2.3.4",
			User:     "test_user",
			Password: "",
			Keyfile:  "/etc/keyfile",
		},
	}
	convertDeviceEnv(devs["device"])
	assert.Equal(t, "1.2.3.4", devs["device"].Address)
	assert.Equal(t, "test_user", devs["device"].User)
	assert.Equal(t, "", devs["device"].Password)
	assert.Equal(t, "/etc/keyfile", devs["device"].Keyfile)

	devs = map[string]*DeviceConfig{
		"device": {
			Address:  "$addr",
			User:     "$user",
			Password: "$password",
			Keyfile:  "$keyfile",
		},
	}
	convertDeviceEnv(devs["device"])

	assert.Equal(t, "address", devs["device"].Address)
	assert.Equal(t, "root", devs["device"].User)
	assert.Equal(t, "", devs["device"].Password)
	assert.Equal(t, "/etc/kf", devs["device"].Keyfile)

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
	assert.ErrorIs(t, err, utils.Err(errFailParseConfig))
	assert.Nil(t, config, "Invalid YAML should not return config")

	config, err = LoadConfig([]byte("{}"))
	assert.Nil(t, err, "No error loading empty config")
	assert.NotNil(t, config, "Empty config is non-nil")
}
