// Package deployment defines the operations for safely replacing a device configuration, including a configurable rollback.
//
// # Deployment lifecycle
//
// This sequence details the deployment process.
// Since configuration is global, concurrent deploys to a single device is not possible.
// While applying changes can be done automatically, the people managing the device must decide whether to accept the result.
// Confirmation can be done directly on the device, or (WIP) using a GitHub action.
//
//	    Upload(Artifact)
//	           |
//	       RemoteFile         a handle to the resulting remote configuration file
//	           |
//	 LoadCandidate(RemoteFile)
//	           |              changes staged, device still unchanged
//	    CommitConfirmed
//	           |              changes are active, rollback timer started
//	ReadActiveConfiguration
//	           |
//	Human checks the result
//	           |
//	    Accept changes?
//	     /           \
//	   yes            no (or human verification does not finish in time)
//	    |              |
//	 Confirm       Leave unconfirmed
//	    |              |
//	  Save           Close
//	    |              |
//	  Close        Device rolls back when its timer expires
//
// # Uploading and applying a new configuration
//
// The file must already be in the device's configuration format, including required firmware metadata.
// Upload sends it unchanged to a private temporary file, returning a RemoteFile handle for later reference.
// Pass the file reference returned by Upload to LoadCandidate in the same deployment session.
// This links loading to a completed upload and ensures the session can remove its temporary files afterward.
// References from another session and incomplete uploads are rejected.
//
// IMPORTANT:
// LoadCandidate fully replaces the staged configuration with the uploaded file, which can delete settings absent from the new version!
//
// Call CommitConfirmed after loading succeeds, using the same session so it applies the configuration just loaded.
// Wait for each step to finish before starting the next, and avoid concurrent configuration changes on the device during verification.
//
// # Checking and accepting changes
//
// The people managing the device should check that the active configuration has the intended settings, the device remains reachable, and the affected services work as intended.
//
// If these checks pass, call Confirm and then Save.
// (Confirm stops rollback, while Save persists the accepted configuration across restart).
// `Confirm` must succeed prior to executing `Save`, and some devices require the two to run in the same session.
//
// ReadActiveConfiguration supplies the active settings for this check.
// Devices may normalize ordering, whitespace, defaults, and metadata, so differences from the uploaded bytes do not necessarily mean deployment failed.
// ConfirmationPending reports whether rollback is armed, but a cleared timer alone cannot tell the human whether changes were confirmed or rolled back.
//
// The timeout allows automatic rollback if connectivity goes down or in other exceptional circumstances, and the configuration is not confirmed within the threshold.
// Note the device's timer continues independently of the client connection or other network channel timeouts.
// Ensure your ConfirmTimeout is a positive duration that allows enough time for you to confidently verify configuration changes (this is customizable per device).
//
// # Errors and cleanup
//
// If any step fails, the whole deployment is stopped and cleanup will remove any loaded files.
// A failed commit can still leave changes active with rollback armed, so check the active configuration and pending timer before retrying or confirming anything.
// If the device is unreachable, leave the deployment unconfirmed and check again after the rollback window.
//
// Note that `Confirm“ will cancel the rollback, so if `Save` fails afterwards then the persistence is uncertain and changes *may or may not* revert on restart.
// Check the running configuration and resolve the save failure before relying on those settings to persist.
//
// Cleanup of temporary uploaded files occurs whenever the session or connection closes.
// No further command execution occurs, so on unclean exits this could result in situations such as:
// * Leftover temporary files
// * The new configuration unconfirmed with a rollback pending
// * The configuration confirmed but not saved
//
// *WARNING*:
// Configuration and device diagnostics can contain credentials, so printable errors omit them and you should avoid logging them without sanitization.
package deployment
