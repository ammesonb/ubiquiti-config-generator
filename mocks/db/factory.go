package db

import (
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New() (db.Service, error) {
	inMemoryDatabaseName := "file::memory:?cache=shared"

	gormDb, err := gorm.Open(sqlite.Open(inMemoryDatabaseName), &gorm.Config{})
	return &db.DefaultDatabaseService{Database: gormDb}, err
}
