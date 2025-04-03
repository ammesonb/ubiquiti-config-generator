package db

import (
	"time"

	"gorm.io/gorm"
)

// CommitCheck contains details about the validity of configurations in a commit
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

// CheckLog contains a single log entry about configuration validity
type CheckLog struct {
	gorm.Model
	Revision    string
	CommitCheck CommitCheck `gorm:"foreignKey:Revision;references:Revision"`
	Status      string
	Timestamp   time.Time
	Message     string
}

// Deployment contains details about a deployment of a configuration
type Deployment struct {
	gorm.Model
	ID           int    `gorm:"primaryKey"`
	FromRevision string `gorm:"deployment_revision"`
	ToRevision   string `gorm:"deployment_revision"`
	Status       string
	StartedAt    time.Time
	EndedAt      time.Time
}

// DeploymentLog contains a single log entry about a deployment of a configuration
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
