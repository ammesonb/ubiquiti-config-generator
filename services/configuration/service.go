package configuration

import (
	"os"
	"reflect"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	yaml "gopkg.in/yaml.v3"
)

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . Service
type Service interface {
	Load(filepath string, fsService filesystem.Service) error
	GetGitConfig() GitConfig
	GetLoggingConfig() LoggingConfig
	GetDevices() []*DeviceConfig
}

// YAMLService provides configuration provided from YAML definitions
type YAMLService struct {
	config *Config
}

// GetGitConfig retrieves the git-specific configuration
func (s *YAMLService) GetGitConfig() GitConfig {
	return s.config.Git
}

// GetLoggingConfig retrieves the logging-specific configuration
func (s *YAMLService) GetLoggingConfig() LoggingConfig {
	return s.config.Logging
}

// GetDevices returns a list of all configured devices
func (s *YAMLService) GetDevices() []*DeviceConfig {
	return s.config.Devices
}

// Load populates this configuration service with data from the provided file
func (s *YAMLService) Load(filepath string, fsService filesystem.Service) error {
	configBytes, err := fsService.ReadFile(filepath)
	if err != nil {
		return errors.ErrWithParent(errReadMainConfig, err)
	}

	s.config = &Config{}
	if err = yaml.Unmarshal(configBytes, s.config); err != nil {
		return errors.ErrWithCtxParent(errParseMainConfig, filepath, err)
	}

	getConfigValuesFromEnv(s.config)
	return loadDevices(s.config.DevicesFile, &s.config.Devices, fsService)
}

func loadDevices(devicesFile string, devices *[]*DeviceConfig, fsService filesystem.Service) error {
	if devicesFile == "" {
		return nil
	}

	devicesBytes, err := fsService.ReadFile(devicesFile)
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
		// NOTE:
		// should this support other types? e.g. ports as ints
		// tricky due to typing conflicts - config value would be string (${ENV_NAME}) but then
		// actual value would have to be another type, which means either allowing `any` value for config,
		// overriding the struct type either for initial environment name detection as string (vs int for port value, e.g.)
		// or typing for the target env value, which would result in the string ENV_NAME breaking unmarshal
		if field.Kind() == reflect.String {
			if shouldGetValueFromEnv(field.String()) {
				field.SetString(os.Getenv(stripEnvMarker(field.String())))
			}
		}
	}
}

// MakeDefaultConfigurationService creates a new default configuration service from the provided config
func MakeDefaultConfigurationService(config *Config) *YAMLService {
	return &YAMLService{config: config}
}
