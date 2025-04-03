package services

import (
	"math/rand"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/stretchr/testify/assert"
)

type mockConfigSvc struct {
	isSingleton bool
}

func (m mockConfigSvc) IsSingleton() bool {
	return m.isSingleton
}

// pretend service is an int, but it could be anything
var svc int = 4

func (m mockConfigSvc) New() (any, error) {
	// For singleton, return by reference otherwise by value
	if m.isSingleton {
		return &svc, nil
	}

	return rand.Int(), nil
}

func TestSingletonRegistration(t *testing.T) {
	RegisterService(ConfigurationServiceIndex, mockConfigSvc{isSingleton: true})
	assert.Contains(t, serviceRegistrations, ConfigurationServiceIndex)

	// Ensure service is created and cached
	fetchedService, err := GetService(ConfigurationServiceIndex)
	assert.NoError(t, err)
	assert.Contains(t, serviceCache, ConfigurationServiceIndex)
	assert.Equal(t, &svc, fetchedService)

	// Ensure on second call, same service is returned and cache is not modified
	fetchedService, err = GetService(ConfigurationServiceIndex)
	assert.NoError(t, err)
	assert.Equal(t, &svc, fetchedService)
	assert.Equal(t, &svc, serviceCache[ConfigurationServiceIndex])
}

func TestEphemeralRegistration(t *testing.T) {
	RegisterService(FilesystemServiceIndex, mockConfigSvc{isSingleton: false})
	assert.Contains(t, serviceRegistrations, FilesystemServiceIndex)
	assert.False(t, serviceRegistrations[FilesystemServiceIndex].IsSingleton())

	// Ensure service is created but not cached
	fetchedService, err := GetService(FilesystemServiceIndex)
	assert.NoError(t, err)
	assert.Nil(t, serviceCache[FilesystemServiceIndex])
	assert.NotEqual(t, &svc, fetchedService)

	// Ensure on second call, new service instance is returned
	secondService, err := GetService(FilesystemServiceIndex)
	assert.NoError(t, err)
	assert.NotEqual(t, &svc, secondService)
	assert.NotEqual(t, fetchedService, secondService)
	assert.NotEqual(t, &fetchedService, &secondService)
}

func TestMissedRegistration(t *testing.T) {
	var serviceName ServiceIndex = "nonexistent"
	_, err := GetService(serviceName)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errNoSuchService, serviceName))
}
