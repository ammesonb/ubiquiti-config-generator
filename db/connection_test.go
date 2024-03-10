package db

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

func getTestDB() (*gorm.DB, error) {
	return OpenDB("file::memory:?cache=shared")
}

func TestOpenDB(t *testing.T) {
	// Try an invalid data source connection string
	db, err := OpenDB("file::invalid")
	assert.Nil(t, db, "No DB for bad connection string")
	assert.ErrorIs(t, err, utils.Err(errConnect))

	// Attempt to establish an in-memory connection for tests
	db, err = getTestDB()
	assert.NotNil(t, db)
	assert.NoError(t, err)

	// Check with actual file, including cleanup
	db, err = OpenDB("test.db")
	assert.NotNil(t, db)
	assert.NoError(t, err)
	defer func() {
		dbActual, err := db.DB()
		if err != nil {
			fmt.Println("Failed to get SQL DB from gorm: " + err.Error())
		} else if err = dbActual.Close(); err != nil {
			fmt.Println("Failed to close SQL DB: " + err.Error())
		} else if err := os.Remove("test.db"); err != nil {
			fmt.Println("Failed to remove test DB: " + err.Error())
		}
	}()
}
