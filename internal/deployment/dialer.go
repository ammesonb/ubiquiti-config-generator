package deployment

import (
	"context"
	"time"
)

// Dialer opens a deployment session on one configured device. Implementations
// own authentication, host identity verification, and firmware-specific behavior.
// Application code must not depend on the testlab package for device access.
type Dialer interface {
	Dial(ctx context.Context) (Session, error)
}

// Session owns a native configuration session and its uploaded files. Methods
// are sequential; callers must exclude conflicting deployments on the same device.
// Every operation must honor context cancellation, including connection setup.
// Failures should wrap their cause in a StageError without exposing configuration
// or device output in the error message.
type Session interface {
	// Upload transfers the complete artifact to a private temporary file.
	Upload(ctx context.Context, artifact Artifact) (RemoteFile, error)

	// Load replaces the candidate, including deleting settings absent from file.
	// Load and CommitConfirmed must use the same native configuration session.
	Load(ctx context.Context, file RemoteFile) error

	// CommitConfirmed applies the candidate and arms native timeout recovery.
	// Nil requires native evidence that the commit succeeded, including checking
	// for remaining candidate changes where exit status is unreliable.
	// It never confirms or saves. An error can leave changes and a timer active;
	// callers must inspect state rather than infer that nothing changed.
	CommitConfirmed(ctx context.Context, timeout time.Duration) error

	// ReadActive retrieves current configuration rather than the staged candidate.
	ReadActive(ctx context.Context) (Configuration, error)

	// ConfirmationPending queries the native rollback mechanism. False with a nil
	// error alone cannot distinguish confirmation from rollback or prove success.
	ConfirmationPending(ctx context.Context) (bool, error)

	// Confirm explicitly accepts pending changes and disables native rollback.
	// It does not persist the configuration across device restart.
	Confirm(ctx context.Context) error

	// Save persists active configuration only after explicit confirmation. The
	// implementation must refuse to save while native confirmation is pending.
	Save(ctx context.Context) error

	// Close removes owned temporary files and tears down the candidate session,
	// including after a failed load. It must preserve any armed rollback mechanism
	// and must never implicitly confirm or save. Cleanup failures are returned.
	Close(ctx context.Context) error
}
