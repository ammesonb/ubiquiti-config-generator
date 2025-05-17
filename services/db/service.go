package db

import (
	"fmt"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/charmbracelet/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Service specifies required methods for database interactions
type Service interface {
	GetDB() *gorm.DB
	Migrate() error
	Exists(model any, idCol string, idVal any) (bool, error)
	CloseDB(logger *log.Logger)
}

// DefaultDatabaseService provides default implementations for DatabaseService
type DefaultDatabaseService struct {
	Database *gorm.DB
}

// GetDB returns the database connection
func (s *DefaultDatabaseService) GetDB() *gorm.DB {
	return s.Database
}

// Migrate runs database migrations
func (s *DefaultDatabaseService) Migrate() error {
	return s.Database.AutoMigrate(&CommitCheck{}, &CheckLog{}, &Deployment{}, &DeploymentLog{})
}

// Exists checks if a record exists in the database with a given ID column and value
func (s *DefaultDatabaseService) Exists(model any, idCol string, idVal any) (bool, error) {
	var exists bool

	err := s.Database.
		Model(model).
		Select("COUNT(*) > 0").
		Where(fmt.Sprintf("%s = ?", idCol), idVal).
		Find(&exists).
		Error

	return exists, err
}

func (s *DefaultDatabaseService) CloseDB(logger *log.Logger) {
	sqlDB, err := s.Database.DB()
	if err != nil {
		logger.Errorf("Failed getting database connection on shutdown: %v", err)
	} else if err = sqlDB.Close(); err != nil {
		logger.Errorf("Failed closing database connection on shutdown: %v", err)
	}
}

func New(config configuration.Service) (Service, error) {
	dbName := config.GetLoggingConfig().DBName
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	return &DefaultDatabaseService{Database: db}, err
}
