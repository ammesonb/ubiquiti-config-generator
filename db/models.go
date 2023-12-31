package db

import (
	"time"

	"gorm.io/gorm"
)

type CommitCheck struct {
	gorm.Model
	Revision  string `gorm:"primaryKey"`
	Status    string
	StartedAt time.Time
	EndedAt   time.Time
}

var (
	StatusInProgress = "in_progress"
	StatusInfo       = "info"
	StatusSuccess    = "success"
	StatusFailure    = "failure"
)

type CheckLog struct {
	gorm.Model
	Revision    string
	CommitCheck CommitCheck `gorm:"foreignKey:Revision;references:Revision"`
	Status      string
	Timestamp   time.Time
	Message     string
}

type Deployment struct {
	gorm.Model
	ID           int    `gorm:"primaryKey"`
	FromRevision string `gorm:"deployment_revision"`
	ToRevision   string `gorm:"deployment_revision"`
	Status       string
	StartedAt    time.Time
	EndedAt      time.Time
}

type DeploymentLog struct {
	gorm.Model
	DeploymentID int
	Deployment   Deployment `gorm:"foreignKey:DeploymentID;references:ID"`
	FromRevision string     `gorm:"index:deployment_log"`
	ToRevision   string     `gorm:"index:deployment_log"`
	Status       string
	Timestamp    time.Time
	Message      string
}

func Exists(logDB *gorm.DB, model interface{}, idCol string, idVal interface{}) (bool, error) {
	var exists bool

	err := logDB.
		Model(model).
		Select("SELECT COUNT(*) > 0").
		Where("%s = ?", idCol, idVal).
		Find(&exists).
		Error

	return exists, err
}
