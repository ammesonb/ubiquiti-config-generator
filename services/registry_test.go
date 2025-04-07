package services

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

type mockConfigSvc struct {
	isSingleton bool
}

func (m mockConfigSvc) IsSingleton() bool {
	return m.isSingleton
}

// pretend service is an int, but it could be anything
type mockSvc struct {
	// struct needs a placeholder field, otherwise same empty memory allocation is reused
	id int
}

func (m mockSvc) StopService(_ *log.Logger) {}

var svc *mockSvc = &mockSvc{id: 1}

func (m mockConfigSvc) New() (ServiceImplementation, error) {
	// For singleton, return by reference otherwise by value
	if m.isSingleton {
		return svc, nil
	}

	return &mockSvc{id: rand.Int()}, nil
}

func TestSingletonRegistration(t *testing.T) {
	RegisterService(ConfigurationServiceIndex, mockConfigSvc{isSingleton: true})
	assert.Contains(t, serviceRegistrations, ConfigurationServiceIndex)

	// Ensure service is created and cached
	fetchedService, err := GetService(ConfigurationServiceIndex)
	assert.NoError(t, err)
	assert.Contains(t, serviceCache, ConfigurationServiceIndex)
	assert.Same(t, fetchedService, svc)

	// Ensure on second call, same service is returned and cache is not modified
	fetchedService, err = GetService(ConfigurationServiceIndex)
	assert.NoError(t, err)
	assert.Same(t, fetchedService, svc)
	assert.Same(t, fetchedService, serviceCache[ConfigurationServiceIndex])
}

func TestEphemeralRegistration(t *testing.T) {
	RegisterService(FilesystemServiceIndex, mockConfigSvc{isSingleton: false})
	assert.Contains(t, serviceRegistrations, FilesystemServiceIndex)
	assert.False(t, serviceRegistrations[FilesystemServiceIndex].IsSingleton())

	// Ensure service is created but not cached
	fetchedService, err := GetService(FilesystemServiceIndex)
	assert.NoError(t, err)
	assert.Nil(t, serviceCache[FilesystemServiceIndex])
	fmt.Printf("Fetched service: %p\n", fetchedService)
	fmt.Printf("Global service: %p\n", svc)

	assert.NotSame(t, fetchedService, svc)

	// Ensure on second call, new service instance is returned
	secondService, err := GetService(FilesystemServiceIndex)
	assert.NoError(t, err)
	assert.NotSame(t, secondService, svc)
	assert.NotSame(t, secondService, fetchedService)
}

func TestMissedRegistration(t *testing.T) {
	var serviceName ServiceIndex = "nonexistent"
	_, err := GetService(serviceName)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errNoSuchService, serviceName))
}
