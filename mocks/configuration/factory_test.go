package configuration

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/stretchr/testify/assert"
)

func TestMockConfiguration(t *testing.T) {
	logConfig := configuration.LoggingConfig{}
	gitConfig := configuration.GitConfig{}
	devices := []*configuration.DeviceConfig{}
	config := &configuration.Config{Logging: logConfig, Git: gitConfig, Devices: devices}
	assert.NoError(t, MockConfiguration(t, config))
	// Ensure the service is registered and can be retrieved
	service := configuration.GetService()
	assert.IsType(t, &configuration.DefaultConfigurationService{}, service)
	assert.Equal(t, logConfig, service.GetLoggingConfig())
	assert.Equal(t, gitConfig, service.GetGitConfig())
	assert.Equal(t, devices, service.GetDevices())
}

func TestMockRegistration(t *testing.T) {
	logConfig := configuration.LoggingConfig{}
	gitConfig := configuration.GitConfig{}
	devices := []*configuration.DeviceConfig{}
	config := &configuration.Config{Logging: logConfig, Git: gitConfig, Devices: devices}
	configRegistration := mockConfigurationRegistration{config: config}

	assert.True(t, configRegistration.IsSingleton())
	svc, err := configRegistration.New()
	assert.NoError(t, err)
	assert.IsType(t, &configuration.DefaultConfigurationService{}, svc)

	configSvc := svc.(configuration.ConfigurationService)

	assert.Equal(t, logConfig, configSvc.GetLoggingConfig())
	assert.Equal(t, gitConfig, configSvc.GetGitConfig())
	assert.Equal(t, devices, configSvc.GetDevices())
}
