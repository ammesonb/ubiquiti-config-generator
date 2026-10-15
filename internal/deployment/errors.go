package deployment

// Stage identifies the operation that failed, independently of device output.
type Stage string

const (
	StageConnect Stage = "connect"
	StageUpload  Stage = "upload"
	StageLoad    Stage = "load"
	StageCommit  Stage = "commit"
	StageInspect Stage = "inspect"
	StageConfirm Stage = "confirm"
	StageSave    Stage = "save"
	StageCleanup Stage = "cleanup"
)

// StageError attributes a failure without printing potentially sensitive device
// diagnostics. Err is retained for errors.Is/errors.As and must not be logged
// without examining it for configuration, credentials, or command output.
// A stage error identifies where execution stopped, not the final device state.
type StageError struct {
	Stage Stage
	Err   error
}

func (e *StageError) Error() string { return "deployment " + string(e.Stage) + " failed" }

func (e *StageError) Unwrap() error { return e.Err }
