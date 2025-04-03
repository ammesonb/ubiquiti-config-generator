package db

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockDatabaseRegistration struct{}

// IsSingleton returns database is singleton
func (m mockDatabaseRegistration) IsSingleton() bool {
	return true
}

// New returns a new database connection
func (m mockDatabaseRegistration) New() (any, error) {
	inMemoryDatabaseName := "file::memory:?cache=shared"

	gormDb, err := gorm.Open(sqlite.Open(inMemoryDatabaseName), &gorm.Config{})
	return db.DefaultDatabaseService{Database: gormDb}, err
}

func MockDatabase(_ *testing.T) error {
	services.RegisterService(services.DatabaseServiceIndex, mockDatabaseRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	dbService, err := services.GetService(services.DatabaseServiceIndex)
	if err != nil {
		return err
	}

	return dbService.(db.DatabaseService).Migrate()
}
