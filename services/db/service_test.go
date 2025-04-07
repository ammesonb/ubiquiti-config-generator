package db

import (
	"testing"
	"time"

	mockConf "github.com/ammesonb/ubiquiti-config-generator/mocks/configuration"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))

	t.Run("count fails before migration", func(t *testing.T) {
		n, err := dbRegistration{}.New()
		dbService := n.(DatabaseService)
		assert.NoError(t, err)
		checkRows := int64(0)
		err = dbService.GetDB().Model(&CommitCheck{}).Count(&checkRows).Error
		assert.Error(t, err)
		assert.Zero(t, checkRows)
	})

	t.Run("count succeeds after migration", func(t *testing.T) {
		assert.NoError(t, RegisterService())
		dbService := GetService()
		assert.IsType(t,
			&DefaultDatabaseService{}, dbService)

		// Only way to check migrate worked is by seeing if a count on a model worked
		checkRows := int64(0)
		assert.NoError(t, dbService.GetDB().Model(&CommitCheck{}).Count(&checkRows).Error)
		assert.Zero(t, checkRows, "No rows inserted yet")
	})
}

func TestExists(t *testing.T) {
	assert.NoError(t, mockConf.MockConfiguration(t, mockConfig))
	assert.NoError(t, RegisterService())

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
