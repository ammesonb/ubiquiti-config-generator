package configuration

import (
	"os"
	"reflect"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	yaml "gopkg.in/yaml.v3"
)

type ConfigurationService interface {
	Load(filepath string) error
	GetGitConfig() GitConfig
	GetLoggingConfig() LoggingConfig
	GetDevices() []*DeviceConfig
}

// DefaultConfigurationService implements ConfigurationService
type DefaultConfigurationService struct {
	config *Config
}

// GetGitConfig retrieves the git-specific configuration
func (s *DefaultConfigurationService) GetGitConfig() GitConfig {
	return s.config.Git
}

// GetLoggingConfig retrieves the logging-specific configuration
func (s *DefaultConfigurationService) GetLoggingConfig() LoggingConfig {
	return s.config.Logging
}

// GetDevices returns a list of all configured devices
func (s *DefaultConfigurationService) GetDevices() []*DeviceConfig {
	return s.config.Devices
}

// Load populates this configuration service with data from the provided file
func (s *DefaultConfigurationService) Load(filepath string) error {
	fsService := filesystem.GetService()
	configBytes, err := fsService.ReadFile(filepath)
	if err != nil {
		return errors.ErrWithParent(errReadMainConfig, err)
	}

	s.config = &Config{}
	if err = yaml.Unmarshal(configBytes, s.config); err != nil {
		return errors.ErrWithCtxParent(errParseMainConfig, filepath, err)
	}

	getConfigValuesFromEnv(s.config)
	return loadDevices(s.config.DevicesFile, &s.config.Devices)
}

func loadDevices(devicesFile string, devices *[]*DeviceConfig) error {
	if devicesFile == "" {
		return nil
	}

	devicesBytes, err := filesystem.GetService().ReadFile(devicesFile)
	if err != nil {
		return errors.ErrWithParent(errReadDevices, err)
	}

	if err = yaml.Unmarshal(devicesBytes, devices); err != nil {
		return errors.ErrWithCtxParent(errParseDevices, devicesFile, err)
	}

	for _, device := range *devices {
		updateConfigFromEnv(device)
	}

	return nil
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

// MakeDefaultConfigurationService creates a new default configuration service from the provided config
func MakeDefaultConfigurationService(config *Config) *DefaultConfigurationService {
	return &DefaultConfigurationService{config: config}
}
