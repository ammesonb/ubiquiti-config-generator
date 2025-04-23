package db

import (
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
	mockConf "github.com/ammesonb/ubiquiti-config-generator/mocks/configuration"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))

	t.Run("count fails before migration", func(t *testing.T) {
		assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))
		n, err := dbRegistration{}.New()
		assert.NoError(t, err)
		dbService := n.(DatabaseService)
		tables, err := dbService.GetDB().Debug().Migrator().GetTables()
		assert.NoError(t, err)
		assert.Empty(t, tables)
		checkRows := int64(0)
		err = dbService.GetDB().Model(&CommitCheck{}).Count(&checkRows).Error
		assert.Error(t, err)
		assert.Zero(t, checkRows)
	})

	t.Run("count succeeds after migration", func(t *testing.T) {
		assert.NoError(t, RegisterService(t.Context()))
		dbService := GetService()
		assert.IsType(t,
			&DefaultDatabaseService{}, dbService)

		tables, err := dbService.GetDB().Migrator().GetTables()
		assert.NoError(t, err)
		assert.NotEmpty(t, tables)
	})
}

func TestExists(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))
	assert.NoError(t, RegisterService(t.Context()))

	dbService := GetService()

	// Create simple data entry
	dbService.GetDB().Create(&CommitCheck{
		Revision:  "abcd",
		Status:    "pending",
		StartedAt: time.Now(),
	})

	// Ensure it exists, and something with a different ID does not
	created, err := dbService.Exists(CommitCheck{}, "Revision", "abcd")
	assert.NoError(t, err)
	assert.True(t, created, "Check created")

	exists, err := dbService.Exists(CommitCheck{}, "Revision", "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists, "Check does not exist")

	exists, err = dbService.Exists(CommitCheck{}, "BadColumn", "abcd")
	assert.Error(t, err)
	assert.False(t, exists, "Check does not exist for bad ID column")
}

func TestStopService(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))
	assert.NoError(t, RegisterService(t.Context()))
	dbService := GetService()
	assert.IsType(t,
		&DefaultDatabaseService{}, dbService)

	dbService.StopService(console_logger.DefaultLogger())
	// Only way to check migrate worked is by seeing if a count on a model worked
	checkRows := int64(0)
	// should error if service is stopped
	// error is private to sql package so cannot check type
	assert.Error(t, dbService.GetDB().Model(&CommitCheck{}).Count(&checkRows).Error)
}
