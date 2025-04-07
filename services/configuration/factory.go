package configuration

import (
	"os"
	"path"

	"github.com/ammesonb/ubiquiti-config-generator/services"
)

var configEnvVar = "UBQ_CONFIGURATION_FILE"
var defaultConfigPath = "./config.yaml"

type configRegistration struct{}

// IsSingleton returns configuration service is singleton
func (configRegistration) IsSingleton() bool {
	return true
}

// New instantiates a new ConfigurationService
func (configRegistration) New() (services.ServiceImplementation, error) {
	var configuration ConfigurationService = &DefaultConfigurationService{}

	err := configuration.Load(getConfigurationPath())
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

// RegisterService registers the configuration service
func RegisterService() error {
	services.RegisterService(services.ConfigurationServiceIndex, configRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.ConfigurationServiceIndex)

	return err
}

// GetService returns the configuration service
func GetService() ConfigurationService {
	svc, _ := services.GetService(services.ConfigurationServiceIndex)

	return svc.(ConfigurationService)
}

func getConfigurationPath() string {
	configPath := os.Getenv(configEnvVar)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	return path.Clean(configPath)
}
