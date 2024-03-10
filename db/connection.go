package db

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	errConnect = "failed to connect to database"
	errMigrate = "failed migrating DB %s table"
)

// for windows, requires TDM-GCC and CGO_ENABLED=1 set
// https://jmeubank.github.io/tdm-gcc/
func OpenDB(dbName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, utils.ErrWithParent(errConnect, err)
	}

	structures := map[string]any{
		"commit check":   &CommitCheck{},
		"check log":      &CheckLog{},
		"deployment":     &Deployment{},
		"deployment log": &DeploymentLog{},
	}

	for name, structure := range structures {
		if err = db.AutoMigrate(structure); err != nil {
			return nil, utils.ErrWithCtxParent(errMigrate, name, err)
		}
	}

	return db, nil
}
