package configuration

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
)

type mockConfigurationRegistration struct {
	config *configuration.Config
}

// IsSingleton returns configuration is singleton
func (m mockConfigurationRegistration) IsSingleton() bool {
	return true
}

// New returns a new mock ConfigurationService
func (m mockConfigurationRegistration) New() (any, error) {
	return configuration.MakeDefaultConfigurationService(m.config), nil
}

// MockConfiguration registers the mocked configuration service, requiring a test object to ensure only used in tests
func MockConfiguration(_ *testing.T, config *configuration.Config) error {
	services.RegisterService(services.ConfigurationServiceIndex, mockConfigurationRegistration{config: config})
	// Attempt get, so we know upfront if the service can load successfully
	_, err := services.GetService(services.ConfigurationServiceIndex)

	return err
}
