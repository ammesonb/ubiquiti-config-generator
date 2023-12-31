package db

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDB(dbName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	if err = db.AutoMigrate(&CommitCheck{}); err != nil {
		return nil, fmt.Errorf("failed migrating commit check structure: %v", err)
	} else if err = db.AutoMigrate(&CheckLog{}); err != nil {
		return nil, fmt.Errorf("failed migrating check log structure: %v", err)
	} else if err = db.AutoMigrate(&Deployment{}); err != nil {
		return nil, fmt.Errorf("failed migrating deployment structure: %v", err)
	} else if err = db.AutoMigrate(&DeploymentLog{}); err != nil {
		return nil, fmt.Errorf("failed migrating deployment log structure: %v", err)
	}

	return db, nil
}
