// Package deployment defines the contracts for deploying complete native router
// configurations.
package deployment

import "time"

// Artifact is a complete native configuration, including any firmware metadata
// required for loading. Content is used as supplied, without templating or merging.
// It can contain credentials and must not be included in logs or errors.
type Artifact struct {
	Content []byte
}

// Request describes one deployment to the device bound to a Dialer.
// The caller must not mutate the artifact while deployment is running.
type Request struct {
	Artifact Artifact

	// ConfirmTimeout is the native rollback window. It must be positive and
	// representable by the device implementation without silently rounding or disabling it.
	// It is independent of the context deadline used for network operations.
	ConfirmTimeout time.Duration
}

// RemoteFile identifies an uploaded artifact on the connected device. It is
// session-owned and must not be constructed from caller-supplied remote paths.
type RemoteFile string

// Configuration is a native serialization of active settings. It may differ from
// the uploaded artifact in ordering, whitespace, defaults, or generated metadata,
// so byte equality with the artifact alone is not a valid deployment check.
// Configuration can contain credentials and must not be logged.
type Configuration []byte
