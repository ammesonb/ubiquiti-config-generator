package db

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDB(dbName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, utils.ErrWithParent("failed to connect database", err)
	}

	if err = db.AutoMigrate(&CommitCheck{}); err != nil {
		return nil, utils.ErrWithParent("failed migrating commit check structure", err)
	} else if err = db.AutoMigrate(&CheckLog{}); err != nil {
		return nil, utils.ErrWithParent("failed migrating check log structure", err)
	} else if err = db.AutoMigrate(&Deployment{}); err != nil {
		return nil, utils.ErrWithParent("failed migrating deployment structure", err)
	} else if err = db.AutoMigrate(&DeploymentLog{}); err != nil {
		return nil, utils.ErrWithParent("failed migrating deployment log structure", err)
	}

	return db, nil
}
