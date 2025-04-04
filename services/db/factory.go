package db

import (
	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dbRegistration struct {
}

// IsSingleton returns database is singleton
func (m dbRegistration) IsSingleton() bool {
	return true
}

// New returns a new database connection
func (m dbRegistration) New() (any, error) {
	dbName := configuration.GetService().GetLoggingConfig().DBName
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	return &DefaultDatabaseService{Database: db}, err
}

// RegisterService registers the database service
func RegisterService() error {
	services.RegisterService(services.DatabaseServiceIndex, dbRegistration{})
	// Attempt get, so we know upfront if the service can load successfully
	dbService, err := services.GetService(services.DatabaseServiceIndex)
	if err != nil {
		return err
	}

	return dbService.(DatabaseService).Migrate()
}

func GetService() DatabaseService {
	dbService, _ := services.GetService(services.DatabaseServiceIndex)
	return dbService.(DatabaseService)
}

// TODO: tests
