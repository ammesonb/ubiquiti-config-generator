package db

import (
	"fmt"

	"gorm.io/gorm"
)

// DatabaseService specifies required methods for database interactions
type DatabaseService interface {
	GetDB() *gorm.DB
	Migrate() error
	Exists(model any, idCol string, idVal any) (bool, error)
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

// TODO: tests
