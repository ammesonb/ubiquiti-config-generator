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

// StageError identifies the failed deployment operation. Its message omits Err
// because the underlying device diagnostics can contain credentials or configuration.
type StageError struct {
	Stage Stage
	Err   error
}

func (e *StageError) Error() string { return "deployment " + string(e.Stage) + " failed" }

func (e *StageError) Unwrap() error { return e.Err }
