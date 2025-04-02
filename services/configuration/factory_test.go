package configuration

import (
	"os"
	"testing"

	mockFs "github.com/ammesonb/ubiquiti-config-generator/mocks/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigurationPath(t *testing.T) {
	t.Run("No env set, using default path", func(t *testing.T) {
		assert.Equal(t, defaultConfigPath, getConfigurationPath())
	})

	t.Run("Env overrides default path", func(t *testing.T) {
		confPath := "/test-config.yaml"
		t.Setenv(configEnvVar, confPath)
		assert.Equal(t, confPath, getConfigurationPath())
	})
}

func TestRegisterService(t *testing.T) {
	assert.NoError(t, RegisterService())

	svc := GetService()
	svcTwo := GetService()

	assert.NotNil(t, svc)
	assert.NotNil(t, svcTwo)
	assert.Equal(t, svc, svcTwo, "Same service interface returned")
	assert.True(t, &svc == &svcTwo, "Same service instance returned")
}

func TestConfigRegistration(t *testing.T) {
	assert.NoError(t, RegisterService())
	assert.NoError(t, filesystem.RegisterService())

	t.Run("service is singleton", func(t *testing.T) {
		assert.True(t, configRegistration{}.IsSingleton())
	})

	t.Run("service instantiates correctly", func(t *testing.T) {
		svc, err := configRegistration{}.New()
		assert.NoError(t, err)
		assert.IsType(t, DefaultConfigurationService{}, svc)
	})

	t.Run("service fails to instantiate", func(t *testing.T) {
		assert.NoError(t, mockFs.MockFileSystem(t))

		mockedFs := filesystem.GetService().(mockFs.MockedFileSystem)
		mockedFs.SetNextResult(mockFs.ReadFileFn, []any{nil, os.ErrNotExist})

		svc, err := configRegistration{}.New()
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, svc)
	})
}
