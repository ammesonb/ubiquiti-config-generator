package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config contains the runtime settings needed to analyze and deploy configurations
type Config struct {
	Devices map[string]DeviceConfig `yaml:"devices"`

	Logging LoggingConfig `yaml:"logging"`

	Git GitConfig `yaml:"git"`
}

type DeviceConfig struct {
	Address  string `yaml:"address"`
	Port     int32  `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Keyfile  string `yaml:"keyfile"`

	TemplatesDir string   `yaml:"templatesDir"`
	ConfigFiles  []string `yaml:"configFiles"`

	CommandFilePath     string `yaml:"command-file-path"`
	ConfigureScriptPath string `yaml:"configure-script-path"`

	AutoRollBack       bool  `yaml:"auto-rollback-on-failure"`
	RebootAfterMinutes int32 `yaml:"reboot-after-minutes"`
	SaveAfterCommit    bool  `yaml:"save-after-commit"`
}

type LoggingConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type GitConfig struct {
	AppId          int32  `yaml:"app-id"`
	PrimaryBranch  string `yaml:"primary-branch"`
	PrivateKeyPath string `yaml:"private-key-path"`
	WebhookURL     string `yaml:"webhook-url"`
	WebhookPort    string `yaml:"webhook-port"`
	WebhookSecret  string `yaml:"webhook-secret"`
}

// ReadConfig takes a path and returns its contents
func ReadConfig(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// LoadConfig takes YAML config data and loads it into the struct
func LoadConfig(config []byte) (*Config, error) {
	conf := &Config{}
	if err := yaml.Unmarshal(config, conf); err != nil {
		return nil, fmt.Errorf("failed loading config: %+v", err)
	}

	convertEnv(conf)

	return conf, nil
}

func shouldGetEnv(name string) bool {
	return strings.HasPrefix(name, "$")
}

func trimYamlEnv(name string) string {
	return strings.TrimLeft(name, "$")
}

func convertEnv(config *Config) {
	for _, device := range config.Devices {

		if shouldGetEnv(device.Address) {
			device.Address = os.Getenv(trimYamlEnv(device.Address))
		}
		if shouldGetEnv(device.User) {
			device.User = os.Getenv(trimYamlEnv(device.User))
		}
		if shouldGetEnv(device.Password) {
			device.Password = os.Getenv(trimYamlEnv(device.Password))
		}
		if shouldGetEnv(device.Keyfile) {
			device.Keyfile = os.Getenv(trimYamlEnv(device.Keyfile))
		}
	}
	if shouldGetEnv(config.Logging.User) {
		config.Logging.User = os.Getenv(trimYamlEnv(config.Logging.User))
	}
	if shouldGetEnv(config.Logging.Password) {
		config.Logging.Password = os.Getenv(trimYamlEnv(config.Logging.Password))
	}
	if shouldGetEnv(config.Git.PrivateKeyPath) {
		config.Git.PrivateKeyPath = os.Getenv(trimYamlEnv(config.Git.PrivateKeyPath))
	}
	if shouldGetEnv(config.Git.WebhookURL) {
		config.Git.WebhookURL = os.Getenv(trimYamlEnv(config.Git.WebhookURL))
	}
	if shouldGetEnv(config.Git.WebhookPort) {
		config.Git.WebhookPort = os.Getenv(trimYamlEnv(config.Git.WebhookPort))
	}
	if shouldGetEnv(config.Git.WebhookSecret) {
		config.Git.WebhookSecret = os.Getenv(trimYamlEnv(config.Git.WebhookSecret))
	}
}
