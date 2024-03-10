package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExists(t *testing.T) {
	// Ensure connection can be opened
	db, err := getTestDB()
	assert.NotNil(t, db)
	assert.NoError(t, err)
	if db == nil {
		t.FailNow()
	}

	// Create simple data entry
	db.Create(&CommitCheck{
		Revision:  "abcd",
		Status:    "pending",
		StartedAt: time.Now(),
	})

	// Ensure it exists, and something with a different ID does not
	created, err := Exists(db, CommitCheck{}, "Revision", "abcd")
	assert.NoError(t, err)
	assert.True(t, created, "Check created")

	exists, err := Exists(db, CommitCheck{}, "Revision", "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists, "Check does not exist")

	exists, err = Exists(db, CommitCheck{}, "BadColumn", "abcd")
	assert.Error(t, err)
	assert.False(t, exists, "Check does not exist for bad ID column")

	// Open a new connection, and make sure the memory DB is shared between them
	db2, err := getTestDB()
	assert.NotNil(t, db2)
	assert.NoError(t, err)

	exists, err = Exists(db, CommitCheck{}, "Revision", "abcd")
	assert.NoError(t, err)
	assert.True(t, exists, "Check still exists on new DB connection")
}
