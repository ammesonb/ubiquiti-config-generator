package db

import (
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	dbSvc, err := New()
	assert.NoError(t, err)
	assert.IsType(t,
		&db.DefaultDatabaseService{}, dbSvc)

	checkRows := int64(0)
	assert.NoError(t, dbSvc.GetDB().Model(&db.CommitCheck{}).Count(&checkRows).Error)
	assert.Zero(t, checkRows, "No rows inserted yet")
}
