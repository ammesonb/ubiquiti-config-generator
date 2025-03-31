package config

import (
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config contains the runtime settings needed to analyze and deploy configurations
type Config struct {
	Logging     LoggingConfig `yaml:"logging"`
	Git         GitConfig     `yaml:"git"`
	DevicesFile string        `yaml:"devices-file"`
	Devices     []DeviceConfig
}

type LoggingConfig struct {
	DBName   string `yaml:"dbConnection"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type GitConfig struct {
	AppID          string `yaml:"app-id"`
	PrimaryBranch  string `yaml:"primary-branch"`
	PrivateKeyPath string `yaml:"private-key-path"`
	WebhookURL     string `yaml:"webhook-url"`
	ListenIP       string `yaml:"listen-ip"`
	WebhookPort    string `yaml:"webhook-port"`
	WebhookSecret  string `yaml:"webhook-secret"`
}

// ReadConfig takes a path and returns its contents
func ReadConfig(path string) ([]byte, error) {
	return os.ReadFile(path)
}

var (
	errFailParseConfig       = "failed parsing config"
	errFailParseDeviceConfig = "failed parsing device config"
)

// LoadConfig takes YAML config data and loads it into the struct
func LoadConfig(config []byte) (*Config, error) {
	conf := &Config{}
	if err := yaml.Unmarshal(config, conf); err != nil {
		return nil, errors.ErrWithParent(errFailParseConfig, err)
	}

	getConfigValuesFromEnv(conf)

	return conf, nil
}

// getConfigValuesFromEnv recursively checks every configuration node for environment variables and
// then replaces the configuration values with the value in the environment vars
func getConfigValuesFromEnv(config *Config) {
	updateConfigFromEnv(config)
	updateConfigFromEnv(&config.Git)
	updateConfigFromEnv(&config.Logging)
}

// If a value starts with a $, it should be read from the environment
func shouldGetValueFromEnv(name string) bool {
	return strings.HasPrefix(name, "$")
}

func stripEnvMarker(name string) string {
	return strings.TrimLeft(name, "$")
}

// updateConfigFromEnv recursively checks every configuration value for environment variables
// and then replaces them with the corresponding value
func updateConfigFromEnv(config interface{}) {
	// get the underlying value for a pointer
	v := reflect.Indirect(reflect.ValueOf(config))

	if v.Kind() != reflect.Struct {
		return
	}

	// If a struct, check every field recursively
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// NOTE: should this support other types? e.g. ports as ints
		//       unsure how to get the $<ENV> name, since unmarshalling would fail
		if field.Kind() == reflect.String {
			if shouldGetValueFromEnv(field.String()) {
				field.SetString(os.Getenv(stripEnvMarker(field.String())))
			}
		}
	}
}
