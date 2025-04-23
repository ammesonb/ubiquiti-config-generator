package db

import (
	"testing"

	mockConf "github.com/ammesonb/ubiquiti-config-generator/mocks/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/stretchr/testify/assert"
)

var mockConfig = &configuration.Config{
	Logging: configuration.LoggingConfig{
		// For tests, do not share cached database
		DBName: "file::memory:?cache=private&mode=memory",
	},
	Git:     configuration.GitConfig{},
	Devices: []*configuration.DeviceConfig{},
}

func TestRegistration(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))
	dbRegistration := dbRegistration{}
	assert.True(t, dbRegistration.IsSingleton())

	svc, err := dbRegistration.New()
	assert.NoError(t, err)
	assert.IsType(t, &DefaultDatabaseService{}, svc)

	dbSvc := svc.(DatabaseService)
	assert.NoError(t, dbSvc.Migrate())
}

func TestRegisterService(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))
	assert.NoError(t, RegisterService(t.Context()))
	dbService := GetService()
	assert.IsType(t,
		&DefaultDatabaseService{}, dbService)

	checkRows := int64(0)
	assert.NoError(t, dbService.GetDB().Model(&CommitCheck{}).Count(&checkRows).Error)
	assert.Zero(t, checkRows, "No rows inserted yet")
}
