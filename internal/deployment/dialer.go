package deployment

import (
	"context"
	"time"
)

// Dialer establishes access to a device for configuration deployment.
type Dialer interface {
	Dial(ctx context.Context) (Session, error)
}

// Session keeps uploaded files and staged configuration together for one deployment.
type Session interface {
	Upload(ctx context.Context, artifact Artifact) (RemoteFile, error)

	LoadCandidate(ctx context.Context, file RemoteFile) error

	CommitConfirmed(ctx context.Context, timeout time.Duration) error

	ReadActiveConfiguration(ctx context.Context) (Configuration, error)

	ConfirmationPending(ctx context.Context) (bool, error)

	Confirm(ctx context.Context) error

	Save(ctx context.Context) error

	Close(ctx context.Context) error
}
