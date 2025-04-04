package db

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/stretchr/testify/assert"
)

func TestMockRegistration(t *testing.T) {
	dbRegistration := mockDatabaseRegistration{}
	assert.True(t, dbRegistration.IsSingleton())

	svc, err := dbRegistration.New()
	assert.NoError(t, err)
	assert.IsType(t, &db.DefaultDatabaseService{}, svc)

	dbSvc := svc.(db.DatabaseService)
	assert.NoError(t, dbSvc.Migrate())
}

func TestMockDatabase(t *testing.T) {
	assert.NoError(t, MockDatabase(t))
	dbService := db.GetService()
	assert.IsType(t,
		&db.DefaultDatabaseService{}, dbService)

	checkRows := int64(0)
	assert.NoError(t, dbService.GetDB().Model(&db.CommitCheck{}).Count(&checkRows).Error)
	assert.Zero(t, checkRows, "No rows inserted yet")
}
