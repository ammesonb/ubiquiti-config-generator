// Package vyos deploys complete configuration files to VyOS devices over SSH.
// Use it to execute the deployment lifecycle described in package deployment, with the VyOS-specific requirements below.
//
// # Preparing a deployment session
//
// Configure Dialer.SSH with the device address, authentication, and trusted host-key verification, then call Dial.
// Use an account that can run the VyOS configuration commands and their required sudo commands without an interactive password prompt.
// Dial connects to the device and prepares its configuration session before returning.
//
//	Dial
//	  |
//	Authenticate over SSH
//	  |
//	Prepare the VyOS configuration session
//	  |
//	Ready to upload and load a configuration
//	  |
//	Keep this session alive throughout commit, verification, confirmation, and save
//	  |
//	Close
//
// Session *preparation* uses SSH.CommandTimeout, defaulting to 30 seconds when zero. This sets up the VyOS-required environment and ensures a persistent session is available throughout the deployment process.
// Note the timeout is separate from SSH.ConnectTimeout, which only covers establishing the SSH connection.
// If preparation fails or never reaches readiness, Dial returns an error and closes the connection.
//
// # Keeping staged changes available
//
// Keep using the session returned by Dial throughout the deployment so later operations act on the configuration you loaded.
// VyOS associates a staged configuration with a running vbash process, so the adapter keeps that vbash session alive between commands.
// (Individual commands still have separate SSH input and output streams, allowing configuration reads without mixing in output from previous commands)
//
// Opening a new session does not resume the previous deployment, so do not reconnect between loading and committing, or between committing and confirming.
// If the connection is lost after a commit, check the device's active configuration and rollback status before starting another deployment!
//
// # Applying and accepting changes
//
// Choose a confirmation window of at least one minute, in whole minutes.
// For example, five minutes is accepted, while 90 seconds is rejected rather than rounded.
// This window controls device rollback, independently of the SSH command timeout.
//
// Resolve the existing deployment by confirming the intended changes or allowing rollback before applying another configuration.
// ReadActiveConfiguration reads the running configuration, leaving the saved startup configuration unchanged.
//
// After verifying the result, call Confirm and then Save through the same session.
// Save writes the accepted configuration to /config/config.boot for use after restart.
// Confirming directly on the device does not enable Save in this adapter session, so if you confirm manually, save on the device as well.
//
// **NOTE:**
// LoadCandidate and CommitConfirmed refuse to proceed while a confirmation timer is already pending.
//
// # Handling failures and closing the session
//
// A successful command exit alone does not establish that VyOS accepted a commit.
// CommitConfirmed also checks that the staged changes were applied and the rollback timer is active, returning an error if either check fails.
// After a commit error, inspect ReadActiveConfiguration and ConfirmationPending before deciding whether to retry.
//
// Errors identify whether connecting, preparing the session, waiting for readiness, or a later deployment step failed.
// For command failures, inspect the wrapped CommandError for stdout, stderr, and the underlying SSH error when you need more diagnostic detail.
//
// **WARNING**:
// Sanitize those diagnostics before logging because they may contain configuration or credentials.
//
// Ensure you call Close when finished, including after a failed upload, load, or commit.
// It attempts to discard staged changes and remove temporary files, and returns an error if cleanup fails.
// An unclean disconnect can leave temporary files behind, so check the device if cleanup cannot complete.
// Closing the session never confirms or saves changes, and leaves any pending device rollback armed.
package vyos
