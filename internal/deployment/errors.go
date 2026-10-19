package deployment

import (
	"context"
	"errors"
)

type Stage string

const (
	StageConnect    Stage = "connect"
	StagePrepare    Stage = "prepare configuration session"
	StageAwaitReady Stage = "wait for session readiness"
	StageUpload     Stage = "upload"
	StageLoad       Stage = "load"
	StageCommit     Stage = "commit"
	StageInspect    Stage = "inspect"
	StageApproval   Stage = "await confirmation"
	StageConfirm    Stage = "confirm"
	StageSave       Stage = "save"
	StageCleanup    Stage = "cleanup"
)

// StageError associates a failure with the deployment step that encountered it.
type StageError struct {
	Stage Stage
	Err   error
}

func (e *StageError) Error() string {
	// Omit the underlying error text because device diagnostics can contain credentials.
	message := "deployment " + string(e.Stage) + " failed"
	if errors.Is(e.Err, context.DeadlineExceeded) {
		return message + ": context deadline exceeded"
	}
	if errors.Is(e.Err, context.Canceled) {
		return message + ": context canceled"
	}
	return message
}

func (e *StageError) Unwrap() error { return e.Err }
