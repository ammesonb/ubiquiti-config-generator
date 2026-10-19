package deployment

import "time"

// Artifact is the replacement configuration file to deploy to a device.
type Artifact struct {
	Content []byte
}

// Request pairs a replacement configuration with the time allowed to verify
// it before the device rolls back unconfirmed changes.
type Request struct {
	Artifact Artifact

	ConfirmTimeout   time.Duration
	SaveAfterConfirm bool
}

// RemoteFile is the handle passed from Upload to LoadCandidate.
type RemoteFile string

// Configuration is a snapshot of the settings currently active on the device.
type Configuration []byte
