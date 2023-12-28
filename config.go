package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config contains the runtime settings needed to analyze and deploy configurations
type Config map[string]RouterConfig

type RouterConfig struct {
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

	Logging LoggingConfig `yaml:"logging"`

	Git GitConfig `yaml:"git"`
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

	for _, router := range *conf {
		convertEnv(&router)
	}

	return conf, nil
}

func shouldGetEnv(name string) bool {
	return strings.HasPrefix(name, "$")
}

func trimYamlEnv(name string) string {
	return strings.TrimLeft(name, "$")
}

func convertEnv(router *RouterConfig) {
	if shouldGetEnv(router.Address) {
		router.Address = os.Getenv(trimYamlEnv(router.Address))
	}
	if shouldGetEnv(router.User) {
		router.User = os.Getenv(trimYamlEnv(router.User))
	}
	if shouldGetEnv(router.Password) {
		router.Password = os.Getenv(trimYamlEnv(router.Password))
	}
	if shouldGetEnv(router.Keyfile) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Keyfile))
	}
	if shouldGetEnv(router.Logging.User) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Logging.User))
	}
	if shouldGetEnv(router.Logging.Password) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Logging.Password))
	}
	if shouldGetEnv(router.Git.PrivateKeyPath) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Git.PrivateKeyPath))
	}
	if shouldGetEnv(router.Git.WebhookURL) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Git.WebhookURL))
	}
	if shouldGetEnv(router.Git.WebhookPort) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Git.WebhookPort))
	}
	if shouldGetEnv(router.Git.WebhookSecret) {
		router.Keyfile = os.Getenv(trimYamlEnv(router.Git.WebhookSecret))
	}
}
